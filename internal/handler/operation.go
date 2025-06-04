package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	"github.com/go-logr/logr"
	"github.com/samber/lo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"slices"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	ctrlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
)

type OperationContextKey struct{}

//go:generate mockgen -destination=./mocks/mock_operation.go -package=mocks github.com/Azure/operation-cache-controller/internal/handler OperationHandlerInterface
type OperationHandlerInterface interface {
	EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error)
	EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error)
	EnsureFinalizerRemoved(ctx context.Context) (reconciler.OperationResult, error)
	EnsureAllAppsAreReady(ctx context.Context) (reconciler.OperationResult, error)
	EnsureAllAppsAreDeleted(ctx context.Context) (reconciler.OperationResult, error)
}

type OperationHandler struct {
	operation *v1alpha1.Operation
	logger    logr.Logger
	client    client.Client
	recorder  record.EventRecorder

	apdutils   ctrlutils.AppDeploymentHelper
	oputils    ctrlutils.OperationHelper
	cacheutils ctrlutils.CacheHelper
}

func NewOperationHandler(ctx context.Context, operation *v1alpha1.Operation, logger logr.Logger, client client.Client, recorder record.EventRecorder) OperationHandlerInterface {
	if operationHandler, ok := ctx.Value(OperationContextKey{}).(OperationHandlerInterface); ok {
		return operationHandler
	}

	return &OperationHandler{
		operation: operation,
		logger:    logger,
		client:    client,
		recorder:  recorder,

		apdutils: ctrlutils.NewAppDeploymentHelper(),
		oputils:  ctrlutils.NewOperationHelper(),
	}
}

func (o *OperationHandler) phaseIn(phases ...string) bool {
	return slices.Contains(phases, o.operation.Status.Phase)
}

func (o *OperationHandler) EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("Operation EnsureNotExpired")
	if len(o.operation.Spec.ExpireAt) == 0 {
		return reconciler.ContinueProcessing()
	}
	if o.phaseIn(v1alpha1.OperationPhaseDeleted, v1alpha1.OperationPhaseDeleting) {
		return reconciler.ContinueProcessing()
	}
	expireTime, err := time.Parse(time.RFC3339, o.operation.Spec.ExpireAt)
	if err != nil {
		o.logger.Error(err, fmt.Sprintf("Failed to parse expire time: %s", o.operation.Spec.ExpireAt))
		o.recorder.Event(o.operation, "Warning", "InvalidExpireTime", "Failed to parse expire time")
		return reconciler.ContinueProcessing()
	}
	if time.Now().Before(expireTime) {
		return reconciler.ContinueProcessing()
	}
	// Expired
	o.logger.Info("deleting expired operation", "expireAt", o.operation.Spec.ExpireAt)
	if err := o.client.Delete(ctx, o.operation, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
		o.logger.Error(err, "Failed to delete expired operation")
		o.recorder.Event(o.operation, "Warning", "DeleteFailed", "Failed to delete expired operation")
		return reconciler.RequeueWithError(err)
	}
	// Stop processing if the operation is deleted
	return reconciler.ContinueProcessing()
}

func (o *OperationHandler) EnsureAllAppsAreReady(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("Operation EnsureAllAppsAreReady")
	if o.phaseIn(v1alpha1.OperationPhaseDeleted, v1alpha1.OperationPhaseDeleting) {
		return reconciler.ContinueProcessing()
	}
	if o.phaseIn(v1alpha1.OperationPhaseEmpty) {
		o.logger.V(1).Info("initializing operation status")
		o.oputils.ClearConditions(o.operation)
		o.operation.Status.OperationID = o.oputils.NewOperationId()
	}
	if o.phaseIn(v1alpha1.OperationPhaseReconciling) {
		err := o.reconcilingApplications(ctx)
		if err != nil {
			o.logger.Error(err, "reconciling applications failed")
			o.recorder.Event(o.operation, "Warning", "ReconcileFailed", "Failed to reconcile deployments")
			return reconciler.RequeueWithError(err)
		}

		o.operation.Status.Phase = v1alpha1.OperationPhaseReconciled
		return reconciler.RequeueOnErrorOrStop(o.client.Status().Update(ctx, o.operation))
	}

	// check the diff between the expected and actual apps, set phase to reconciling and requeue if changes
	expectedCacheKey := o.cacheutils.NewCacheKeyFromApplications(o.operation.Spec.Applications)
	if o.operation.Status.CacheKey != expectedCacheKey {
		o.operation.Status.CacheKey = expectedCacheKey
		o.operation.Status.Phase = v1alpha1.OperationPhaseReconciling
	}
	return reconciler.RequeueOnErrorOrContinue(o.client.Status().Update(ctx, o.operation))
}

func (o *OperationHandler) EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("operation EnsureFinalizer")
	if o.operation.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(o.operation, v1alpha1.OperationFinalizerName) {
		controllerutil.AddFinalizer(o.operation, v1alpha1.OperationFinalizerName)
	}
	return reconciler.RequeueOnErrorOrContinue(o.client.Update(ctx, o.operation))
}

func (o *OperationHandler) EnsureFinalizerRemoved(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("operation EnsureFinalizerDeleted")
	if !o.operation.ObjectMeta.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(o.operation, v1alpha1.OperationFinalizerName) {
		if o.phaseIn(v1alpha1.OperationPhaseDeleted) {
			o.logger.V(1).Info("All app deleted removing finalizer")
			controllerutil.RemoveFinalizer(o.operation, v1alpha1.OperationFinalizerName)
			return reconciler.RequeueOnErrorOrContinue(o.client.Update(ctx, o.operation))
		}
		if !o.phaseIn(v1alpha1.OperationPhaseDeleting) {
			o.logger.V(1).Info("App is not deleted yet, setting phase to deleting")
			o.operation.Status.Phase = v1alpha1.OperationPhaseDeleting
			return reconciler.RequeueOnErrorOrContinue(o.client.Status().Update(ctx, o.operation))
		}
	}
	return reconciler.ContinueProcessing()
}

func (o *OperationHandler) EnsureAllAppsAreDeleted(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("Operation EnsureAllAppsAreDeleted")
	if o.phaseIn(v1alpha1.OperationPhaseDeleting) {
		// deleting logic here
		o.operation.Status.Phase = v1alpha1.OperationPhaseDeleted
		return reconciler.RequeueOnErrorOrStop(o.client.Status().Update(ctx, o.operation))
	}
	return reconciler.ContinueProcessing()
}

func (o *OperationHandler) reconcilingApplications(ctx context.Context) error {
	logger := o.logger.WithValues("operation", "reconcilingApplications")
	currentAppDeployments, err := o.listCurrentAppDeployments(ctx)
	if err != nil {
		return fmt.Errorf("failed to list current appDeployments: %w", err)
	}
	logger.V(1).Info(fmt.Sprintf("current app deployments count %d", len(currentAppDeployments)))
	for _, app := range currentAppDeployments {
		logger.V(1).Info("current app deployment", "appName", app.Name, "opId", app.Spec.OpId, "provision", app.Spec.Provision, "teardown", app.Spec.Teardown, "dependencies", app.Spec.Dependencies)
	}

	expectedAppDeployments := o.expectedAppDeployments()
	logger.V(1).Info(fmt.Sprintf("expected app deployments count %d", len(expectedAppDeployments)))
	for _, app := range expectedAppDeployments {
		logger.V(1).Info("expected app deployment", "appName", app.Name, "opId", app.Spec.OpId, "provision", app.Spec.Provision, "teardown", app.Spec.Teardown, "dependencies", app.Spec.Dependencies)
	}

	added, removed, updated := o.oputils.DiffAppDeployments(expectedAppDeployments, currentAppDeployments, o.oputils.CompareProvisionJobs)
	for _, app := range added {
		logger.V(1).Info(fmt.Sprintf("app to be added %s", app.Name), "opId", app.Spec.OpId, "provision", app.Spec.Provision, "teardown", app.Spec.Teardown, "dependencies", app.Spec.Dependencies)
		if err := ctrl.SetControllerReference(o.operation, &app, o.client.Scheme()); err != nil {
			return fmt.Errorf("failed to set controller reference: %w", err)
		}
		if err := o.client.Create(ctx, &app); err != nil {
			return fmt.Errorf("failed to create app deployment: %w", err)
		}
	}

	for _, app := range removed {
		logger.V(1).Info(fmt.Sprintf("app to be removed %s", app.Name), "opId", app.Spec.OpId, "provision", app.Spec.Provision, "teardown", app.Spec.Teardown, "dependencies", app.Spec.Dependencies)
		if err := o.client.Delete(ctx, &app, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete app deployment: %w", err)
		}
	}

	for _, app := range updated {
		logger.V(1).Info(fmt.Sprintf("app to be updated %s", app.Name), "appName", app.Name, "opId", app.Spec.OpId, "provision", app.Spec.Provision, "teardown", app.Spec.Teardown, "dependencies", app.Spec.Dependencies)
		if err := o.client.Update(ctx, &app); err != nil {
			return fmt.Errorf("failed to update app deployment: %w", err)
		}
	}

	// check if all expected app deployments are ready
	for _, app := range expectedAppDeployments {
		appdeployment := &v1alpha1.AppDeployment{}
		if err := o.client.Get(ctx, client.ObjectKey{Namespace: app.Namespace, Name: app.Name}, appdeployment); err != nil {
			return fmt.Errorf("failed to get app deployment: %w", err)
		}
		// check if all dependencies are ready
		if appdeployment.Status.Phase != v1alpha1.AppDeploymentPhaseReady {
			return fmt.Errorf("app deployment is not ready: name %s, status, %s", app.Name, app.Status.Phase)
		}
	}

	return nil
}

func (o *OperationHandler) expectedAppDeployments() []v1alpha1.AppDeployment {
	return lo.Map(o.operation.Spec.Applications, func(app v1alpha1.ApplicationSpec, index int) v1alpha1.AppDeployment {
		return v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ctrlutils.OperationScopedAppDeployment(app.Name, o.operation.Status.OperationID),
				Namespace: o.operation.Namespace,
			},
			Spec: v1alpha1.AppDeploymentSpec{
				OpId:         o.operation.Status.OperationID,
				Provision:    app.Provision,
				Teardown:     app.Teardown,
				Dependencies: app.Dependencies,
			},
		}
	})
}

func (o *OperationHandler) listCurrentAppDeployments(ctx context.Context) ([]v1alpha1.AppDeployment, error) {
	appDeploymentList := &v1alpha1.AppDeploymentList{}
	if err := o.client.List(ctx, appDeploymentList, client.MatchingFields{v1alpha1.OperationOwnerKey: o.operation.Name}); err != nil {
		return nil, fmt.Errorf("failed to list appDeployments: %w", err)
	}
	return lo.Map(appDeploymentList.Items, func(app v1alpha1.AppDeployment, index int) v1alpha1.AppDeployment {
		return v1alpha1.AppDeployment{
			ObjectMeta: app.ObjectMeta,
			Spec:       app.Spec,
		}
	}), nil
}
