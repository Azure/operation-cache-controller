package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1alpha1 "github.com/Azure/operation-cache-controller/api/v1alpha1"
	ctrlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
	oputils "github.com/Azure/operation-cache-controller/internal/utils/controller/operation"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

//go:generate mockgen -destination=./mocks/mock_cache_adapter.go -package=mocks github.com/Azure/operation-cache-controller/internal/controller CacheAdapterInterface
type CacheAdapterInterface interface {
	CheckCacheExpiry(ctx context.Context) (reconciler.OperationResult, error)
	EnsureCacheInitialized(ctx context.Context) (reconciler.OperationResult, error)
	CalculateKeepAliveCount(ctx context.Context) (reconciler.OperationResult, error)
	AdjustCache(ctx context.Context) (reconciler.OperationResult, error)
}

type CacheAdapter struct {
	cache                      *v1alpha1.Cache
	logger                     logr.Logger
	client                     client.Client
	scheme                     *runtime.Scheme
	recorder                   record.EventRecorder
	setControllerReferenceFunc func(owner, controlled metav1.Object, scheme *runtime.Scheme, opts ...controllerutil.OwnerReferenceOption) error
}

func NewCacheAdapter(ctx context.Context,
	cache *v1alpha1.Cache, logger logr.Logger, client client.Client, scheme *runtime.Scheme, recorder record.EventRecorder,
	fn func(owner, controlled metav1.Object, scheme *runtime.Scheme, opts ...controllerutil.OwnerReferenceOption) error) CacheAdapterInterface {
	return &CacheAdapter{
		cache:                      cache,
		logger:                     logger,
		client:                     client,
		scheme:                     scheme,
		recorder:                   recorder,
		setControllerReferenceFunc: fn,
	}
}

// updateStatus updates the status of the cache cr
func (c *CacheAdapter) updateStatus(ctx context.Context) error {
	if err := c.client.Status().Update(ctx, c.cache); err != nil {
		return fmt.Errorf("unable to update cache status: %w", err)
	}
	return nil
}

// CheckCacheExpiry checks if the cache cr is expired. If it is, the cr is deleted.
func (c *CacheAdapter) CheckCacheExpiry(ctx context.Context) (reconciler.OperationResult, error) {
	if c.cache.Spec.ExpireTime == "" {
		return reconciler.ContinueProcessing()
	}
	ce, err := time.Parse(time.RFC3339, c.cache.Spec.ExpireTime)
	if err != nil {
		c.logger.Error(err, "failed to parse expire time")
		// TODO: set cache expiry condition if needed
		return reconciler.ContinueProcessing()
	}
	if time.Now().After(ce) {
		c.logger.Info("cache is expired, deleting cache cr")
		if err := c.client.Delete(ctx, c.cache); err != nil {
			return reconciler.RequeueWithError(err)
		}
		return reconciler.StopProcessing()
	}
	return reconciler.ContinueProcessing()
}

// EnsureCacheInitialized ensures the cache cr is initialized
func (c *CacheAdapter) EnsureCacheInitialized(ctx context.Context) (reconciler.OperationResult, error) {
	// initialize the AvailableCaches in status if it is nil
	if c.cache.Status.AvailableCaches == nil {
		c.cache.Status.AvailableCaches = []string{}
	}
	if c.cache.Status.CacheKey == "" {
		c.cache.Status.CacheKey = ctrlutils.NewCacheKeyFromApplications(c.cache.Spec.OperationTemplate.Applications)
	}

	return reconciler.RequeueOnErrorOrContinue(c.updateStatus(ctx))
}

// CalculateKeepAliveCount calculates the keepAliveCount for the cache cr
func (c *CacheAdapter) CalculateKeepAliveCount(ctx context.Context) (reconciler.OperationResult, error) {
	// before we have cache service to provide the keepAliveCount, we use fixed value
	c.cache.Status.KeepAliveCount = 5
	return reconciler.RequeueOnErrorOrContinue(c.updateStatus(ctx))
}

func (c *CacheAdapter) createOperationsAsync(ctx context.Context, ops []*v1alpha1.Operation) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error, len(ops))
	for _, op := range ops {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- c.client.Create(ctx, op)
		}()
	}
	wg.Wait()
	close(errChan)
	var errs error
	for err := range errChan {
		errs = errors.Join(errs, err)
	}
	return errs
}

func (c *CacheAdapter) deleteOperationsAsync(ctx context.Context, ops []*v1alpha1.Operation) error {
	wg := sync.WaitGroup{}
	errChan := make(chan error, len(ops))
	for _, op := range ops {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errChan <- c.client.Delete(ctx, op)
		}()
	}
	wg.Wait()
	close(errChan)
	var errs error
	for err := range errChan {
		errs = errors.Join(errs, err)
	}
	return errs
}

func operationReady(op *v1alpha1.Operation) bool {
	return op.Status.Phase == oputils.PhaseReconciled
}

func (c *CacheAdapter) initOperationFromCache(operationName string) *v1alpha1.Operation {
	op := &v1alpha1.Operation{}

	annotations := op.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[ctrlutils.AnnotationNameCacheMode] = ctrlutils.AnnotationValueTrue

	labels := op.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	// TODO: set up requirement label instead
	cacheKeyLabelValue := c.cache.Status.CacheKey
	if len(c.cache.Status.CacheKey) > 63 {
		cacheKeyLabelValue = cacheKeyLabelValue[:63]
	}
	labels[ctrlutils.LabelNameCacheKey] = cacheKeyLabelValue

	op.SetAnnotations(annotations)
	op.SetNamespace(c.cache.Namespace)
	op.SetName(operationName)
	op.SetLabels(labels)
	op.Spec = c.cache.Spec.OperationTemplate
	return op
}

func (c *CacheAdapter) AdjustCache(ctx context.Context) (reconciler.OperationResult, error) {
	var ownedOps v1alpha1.OperationList
	if err := c.client.List(ctx, &ownedOps, client.InNamespace(c.cache.Namespace), client.MatchingFields{cacheOwnerKey: c.cache.Name}); err != nil {
		return reconciler.RequeueWithError(err)
	}
	availableCaches := []string{}
	for _, op := range ownedOps.Items {
		if operationReady(&op) {
			availableCaches = append(availableCaches, op.Name)
		}
	}
	c.cache.Status.AvailableCaches = availableCaches

	keepAliveCount := int(c.cache.Status.KeepAliveCount)
	cacheBalance := len(availableCaches) - keepAliveCount
	switch {
	case cacheBalance == 0:
		// do nothing: should we remove the not available operations?
	case cacheBalance > 0:
		// remove all the not available operations and cut available operations down to keepAliveCount
		availableCacheNumToRemove := cacheBalance
		opsToRemove := []*v1alpha1.Operation{}
		for _, op := range ownedOps.Items {
			if !operationReady(&op) {
				opsToRemove = append(opsToRemove, &op)
			} else {
				if availableCacheNumToRemove > 0 {
					opsToRemove = append(opsToRemove, &op)
					availableCacheNumToRemove--
				}
			}
		}
		c.logger.Info("removing operations", "operations", opsToRemove)
		if err := c.deleteOperationsAsync(ctx, opsToRemove); err != nil {
			return reconciler.RequeueWithError(err)
		}
	case cacheBalance < 0:
		if len(ownedOps.Items) < keepAliveCount {
			// also count not available operations, create new operations to meet the keepAliveCount
			opsToCreate := []*v1alpha1.Operation{}
			opsNumToCreate := keepAliveCount - len(ownedOps.Items)
			for range opsNumToCreate {
				opName := fmt.Sprintf("cached-operation-%s-%s", c.cache.Status.CacheKey[:8], strings.ToLower(ctrlutils.GenerateRandomString(5)))
				opToCreate := c.initOperationFromCache(opName)
				if err := c.setControllerReferenceFunc(c.cache, opToCreate, c.scheme); err != nil {
					return reconciler.RequeueWithError(err)
				}
				opsToCreate = append(opsToCreate, opToCreate)
			}
			c.logger.Info("creating operations", "operations", opsToCreate)
			if err := c.createOperationsAsync(ctx, opsToCreate); err != nil {
				return reconciler.RequeueWithError(err)
			}
		}
		// else do nothing: we assume that any not ready operations are in progress and will be ready
		// we can bring in stuck operations handling if we consider that's one case for cache controller to solve
	}
	return reconciler.RequeueOnErrorOrContinue(c.updateStatus(ctx))
}
