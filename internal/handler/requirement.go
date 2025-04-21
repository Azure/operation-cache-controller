package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	ctlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
	"github.com/Azure/operation-cache-controller/internal/utils/ptr"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

type RequiremenContextKey struct{}

//go:generate mockgen -destination=./mocks/mock_requirement.go -package=mocks github.com/Azure/operation-cache-controller/internal/handler RequirementHandlerInterface
type RequirementHandlerInterface interface {
	EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error)
	EnsureInitialized(ctx context.Context) (reconciler.OperationResult, error)
	EnsureCacheExisted(ctx context.Context) (reconciler.OperationResult, error)
	EnsureCachedOperationAcquired(ctx context.Context) (reconciler.OperationResult, error)
	EnsureOperationReady(ctx context.Context) (reconciler.OperationResult, error)
}

type RequirementHandler struct {
	requirement *v1alpha1.Requirement
	logger      logr.Logger
	client      client.Client
	recorder    record.EventRecorder

	cacheutils ctlutils.CacheHelper
	oputils    ctlutils.OperationHelper
	rqutils    ctlutils.RequirementHelper
}

func NewRequirementHandler(ctx context.Context, requirement *v1alpha1.Requirement, logger logr.Logger, client client.Client, recorder record.EventRecorder) RequirementHandlerInterface {
	if requirementHandler, ok := ctx.Value(RequiremenContextKey{}).(RequirementHandlerInterface); ok {
		return requirementHandler
	}

	return &RequirementHandler{
		requirement: requirement,
		logger:      logger,
		client:      client,
		recorder:    recorder,

		cacheutils: ctlutils.NewCacheHelper(),
		oputils:    ctlutils.NewOperationHelper(),
		rqutils:    ctlutils.NewRequirementHelper(),
	}
}

func (o *RequirementHandler) phaseIn(phases ...string) bool {

	for _, phase := range phases {
		if phase == o.requirement.Status.Phase {
			return true
		}
	}
	return false
}

func (r *RequirementHandler) EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error) {
	r.logger.V(1).Info("operation: EnsureNotExpired")
	if len(r.requirement.Spec.ExpireAt) == 0 {
		return reconciler.ContinueProcessing()
	}

	expireTime, err := time.Parse(time.RFC3339, r.requirement.Spec.ExpireAt)
	if err != nil {
		r.logger.Error(err, fmt.Sprintf("Failed to parse expire time: %s", r.requirement.Spec.ExpireAt))
		r.recorder.Event(r.requirement, "Warning", "InvalidExpireTime", "Failed to parse expire time")
		return reconciler.ContinueProcessing()
	}
	if time.Now().Before(expireTime) {
		return reconciler.ContinueProcessing()
	}
	// Expired
	r.logger.Info("deleting expired requirement", "expireAt", r.requirement.Spec.ExpireAt)
	if err := r.client.Delete(ctx, r.requirement, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
		r.logger.Error(err, "Failed to delete expired requirement")
		r.recorder.Event(r.requirement, "Warning", "DeleteFailed", "Failed to delete expired requirement")
		return reconciler.RequeueWithError(err)
	}
	return reconciler.ContinueProcessing()
}

func (r *RequirementHandler) EnsureInitialized(ctx context.Context) (reconciler.OperationResult, error) {
	r.logger.V(1).Info("operation: EnsureInitialized")
	if !r.phaseIn(v1alpha1.RequirementPhaseEmpty) {
		return reconciler.ContinueProcessing()
	}
	r.requirement.Status.CacheKey = r.cacheutils.NewCacheKeyFromApplications(r.requirement.Spec.Template.Applications)
	r.rqutils.ClearConditions(r.requirement)
	if r.requirement.Spec.EnableCache {
		r.requirement.Status.Phase = v1alpha1.RequirementPhaseCacheChecking
	} else {
		r.requirement.Status.Phase = v1alpha1.RequirementPhaseOperating
	}
	return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
}

func (r *RequirementHandler) ownerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         r.requirement.APIVersion,
		Kind:               r.requirement.Kind,
		Name:               r.requirement.Name,
		UID:                r.requirement.UID,
		Controller:         ptr.Of(true),
		BlockOwnerDeletion: ptr.Of(true),
	}
}

func (r *RequirementHandler) setCacheNotExistedStatus() {
	r.requirement.Status.Phase = v1alpha1.RequirementPhaseOperating
	_ = r.rqutils.UpdateCondition(r.requirement, v1alpha1.RequirementConditionCacheResourceFound, metav1.ConditionFalse, v1alpha1.RequirementConditionReasonCacheCRFound, "Cache CR found")
}

func (r *RequirementHandler) setCacheHitStatus() {
	r.requirement.Status.Phase = v1alpha1.RequirementPhaseReady
	_ = r.rqutils.UpdateCondition(r.requirement, v1alpha1.RequirementConditionOperationReady, metav1.ConditionTrue, v1alpha1.RequirementConditionReasonCacheHit, "Cached Operation acquired")
}

func (r *RequirementHandler) setCacheMissStatus() {
	r.requirement.Status.Phase = v1alpha1.RequirementPhaseOperating
	_ = r.rqutils.UpdateCondition(r.requirement, v1alpha1.RequirementConditionCachedOperationAcquired, metav1.ConditionTrue, v1alpha1.RequirementConditionReasonCacheMiss, "No cached operation available")
}

func (r *RequirementHandler) defaultCacheName() string {
	return fmt.Sprintf("cache-%s", r.requirement.Status.CacheKey)
}

func (r *RequirementHandler) EnsureCacheExisted(ctx context.Context) (reconciler.OperationResult, error) {
	if !r.phaseIn(v1alpha1.RequirementPhaseCacheChecking) {
		return reconciler.ContinueProcessing()
	}

	r.logger.V(1).Info("operation: EnsureCacheExisted")

	// candidate operation id exists, go to next step to acquire the operation
	if len(r.requirement.Status.OperationName) != 0 {
		return reconciler.ContinueProcessing()
	}

	if len(r.requirement.Status.CacheKey) == 0 {
		r.logger.Error(fmt.Errorf("empty cache key"), "Cache key is empty, cannot proceed with cache creation")
		return reconciler.RequeueWithError(fmt.Errorf("empty cache key"))
	}
	cache := &v1alpha1.Cache{}
	// Try to get the Cache CR
	if err := r.client.Get(ctx, types.NamespacedName{Name: r.defaultCacheName(), Namespace: r.requirement.Namespace}, cache); err != nil {
		if client.IgnoreNotFound(err) != nil {
			// If the error is not a NotFound error, return it
			return reconciler.RequeueWithError(err)
		}
		// cache cr not found, create it
		cache.Name = r.defaultCacheName()
		cache.Namespace = r.requirement.Namespace
		cache.Spec = v1alpha1.CacheSpec{
			OperationTemplate: r.requirement.Spec.Template,
			ExpireTime:        r.cacheutils.DefaultCacheExpireTime(),
		}
		err = r.client.Create(ctx, cache)
		if err != nil {
			return reconciler.RequeueWithError(err)
		}
		r.setCacheNotExistedStatus()
		return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
	}
	// extend cache expire time every time when cache is checked
	cache.Spec.ExpireTime = r.cacheutils.DefaultCacheExpireTime()
	_ = r.client.Update(ctx, cache)
	r.requirement.Status.OperationName = r.cacheutils.RandomSelectCachedOperation(cache)
	return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
}

func (r *RequirementHandler) EnsureCachedOperationAcquired(ctx context.Context) (reconciler.OperationResult, error) {
	if !r.phaseIn(v1alpha1.RequirementPhaseCacheChecking) {
		return reconciler.ContinueProcessing()
	}
	r.logger.V(1).Info("operation: EnsureCachedOperationAcquired")
	if len(r.requirement.Status.OperationName) == 0 {
		r.logger.V(1).Info("no cached operation available")
		r.setCacheMissStatus()
		return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
	}
	operation := &v1alpha1.Operation{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: r.requirement.Status.OperationName, Namespace: r.requirement.Namespace}, operation); err != nil {
		r.setCacheMissStatus()
		return reconciler.RequeueOnErrorOrContinue(fmt.Errorf("failed to get operation %s: %w", r.requirement.Status.OperationName, err))
	}
	// already acquired
	if _, ok := operation.Annotations[v1alpha1.OperationAcquiredAnnotationKey]; ok {
		if len(operation.OwnerReferences) != 0 {
			if operation.OwnerReferences[0].UID != r.requirement.UID {
				// return error if owner is not this requirement
				r.logger.V(1).Info("operation already acquired by other requirement", "operation", r.requirement.Status.OperationName)
				r.setCacheMissStatus()
				return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
			} else {
				// set to ready status if the operation already acquired by this requirement
				r.logger.V(1).Info("operation already acquired by this requirement", "operation", r.requirement.Status.OperationName)
				r.setCacheHitStatus()
				return reconciler.RequeueOnErrorOrStop(r.client.Status().Update(ctx, r.requirement))
			}
		}
	}
	// if operation not acquired, acquire it
	if err := r.acquireCachedOperation(ctx, operation); err != nil {
		r.setCacheMissStatus()
		return reconciler.RequeueOnErrorOrContinue(fmt.Errorf("failed to update operation %s: %w", r.requirement.Status.OperationName, err))
	}
	// set to ready status if the operation acquired
	r.setCacheHitStatus()
	return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
}

func (r *RequirementHandler) acquireCachedOperation(ctx context.Context, operation *v1alpha1.Operation) error {
	operation.Annotations[v1alpha1.OperationAcquiredAnnotationKey] = time.Now().Format(time.RFC3339)
	operation.OwnerReferences = []metav1.OwnerReference{r.ownerReference()}
	return r.client.Update(ctx, operation)
}

func (r *RequirementHandler) getOperation() (*v1alpha1.Operation, error) {
	namespacedName := types.NamespacedName{
		Name:      r.requirement.Status.OperationName,
		Namespace: r.requirement.Namespace,
	}

	operation := &v1alpha1.Operation{}
	if err := r.client.Get(context.Background(), namespacedName, operation); err != nil {
		return nil, fmt.Errorf("failed to get operation %s: %w", r.requirement.Status.OperationName, err)
	}
	return operation, nil
}

func (r *RequirementHandler) updateOperation() error {
	op, err := r.getOperation()
	if err != nil {
		return err
	}
	op.Spec = r.requirement.Spec.Template
	return r.client.Update(context.Background(), op)
}

func (r *RequirementHandler) createOperation() error {
	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.requirement.Status.OperationName,
			Namespace: r.requirement.Namespace,
		},
		Spec: r.requirement.Spec.Template,
	}
	if err := controllerutil.SetControllerReference(r.requirement, operation, r.client.Scheme()); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}
	return r.client.Create(context.Background(), operation)
}

func (r *RequirementHandler) EnsureOperationReady(ctx context.Context) (reconciler.OperationResult, error) {
	r.logger.V(1).Info("operation: EnsureOperationReady")
	if r.phaseIn(v1alpha1.RequirementPhaseReady) {
		// check if application changed
		cacheKey := r.cacheutils.NewCacheKeyFromApplications(r.requirement.Spec.Template.Applications)
		if r.requirement.Status.CacheKey != cacheKey {
			r.logger.Info("application changed, updating operation", "oldCacheKey", r.requirement.Status.CacheKey, "newCacheKey", cacheKey)
			if err := r.updateOperation(); err != nil {
				return reconciler.RequeueWithError(err)
			}
			r.requirement.Status.CacheKey = cacheKey
			r.requirement.Status.Phase = v1alpha1.RequirementPhaseOperating
			return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
		}
		return reconciler.ContinueProcessing()
	}
	if !r.phaseIn(v1alpha1.RequirementPhaseOperating) {
		return reconciler.ContinueProcessing()
	}
	if r.rqutils.IsCacheMissed(r.requirement) {
		r.logger.V(1).Info("cache missed, creating operation")
		r.requirement.Status.OperationName = r.requirement.Name + "-" + "operation"
	}
	// check operation status
	if op, err := r.getOperation(); err == nil {
		r.logger.V(1).Info("requirement operation found", "operation", op.Name)
		if op.Status.Phase == v1alpha1.OperationPhaseReconciled {
			r.logger.Info("operation is reconciled, set requirement to ready", "operationName", op.Name, "operationId", op.Status.OperationID)
			r.requirement.Status.Phase = v1alpha1.RequirementPhaseReady
			r.requirement.Status.OperationId = op.Status.OperationID
			return reconciler.RequeueOnErrorOrContinue(r.client.Status().Update(ctx, r.requirement))
		}
		r.logger.V(1).Info("reconciling requirement operation...", "operation", op.Name)
		return reconciler.Requeue()
	}
	r.logger.V(1).Info("operation not found, creating one")
	if err := r.createOperation(); err != nil {
		return reconciler.RequeueWithError(err)
	}
	return reconciler.Requeue()
}
