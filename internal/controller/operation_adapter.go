package controller

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

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	ctlrutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
	apdutils "github.com/Azure/operation-cache-controller/internal/utils/controller/appdeployment"
	oputils "github.com/Azure/operation-cache-controller/internal/utils/controller/operation"
)

type operationAdapterContextKey struct{}

//go:generate mockgen -destination=./mocks/mock_operation_adapter.go -package=mocks github.com/Azure/operation-cache-controller/internal/controller OperationAdapterInterface
type OperationAdapterInterface interface {
	EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error)
	EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error)
	EnsureFinalizerRemoved(ctx context.Context) (reconciler.OperationResult, error)
	EnsureAllAppsAreReady(ctx context.Context) (reconciler.OperationResult, error)
	EnsureAllAppsAreDeleted(ctx context.Context) (reconciler.OperationResult, error)
}

type OperationAdapter struct {
	operation *v1alpha1.Operation
	logger    logr.Logger
	client    client.Client
	recorder  record.EventRecorder
}

func NewOperationAdapter(ctx context.Context, operation *v1alpha1.Operation, logger logr.Logger, client client.Client, recorder record.EventRecorder) OperationAdapterInterface {
	if operationAdapter, ok := ctx.Value(operationAdapterContextKey{}).(OperationAdapterInterface); ok {
		return operationAdapter
	}

	return &OperationAdapter{
		operation: operation,
		logger:    logger,
		client:    client,
		recorder:  recorder,
	}
}

func (o *OperationAdapter) phaseIn(phases ...string) bool {

	for _, phase := range phases {
		if phase == o.operation.Status.Phase {
			return true
		}
	}
	return false
}

func (o *OperationAdapter) EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("Operation EnsureNotExpired")
	if len(o.operation.Spec.ExpireAt) == 0 {
		return reconciler.ContinueProcessing()
	}
	if o.phaseIn(oputils.PhaseDeleted, oputils.PhaseDeleting) {
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

func (o *OperationAdapter) EnsureAllAppsAreReady(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("Operation EnsureAllAppsAreReady")
	if o.phaseIn(oputils.PhaseDeleted, oputils.PhaseDeleting) {
		return reconciler.ContinueProcessing()
	}
	if o.phaseIn(oputils.PhaseEmpty) {
		o.logger.V(1).Info("initializing operation status")
		oputils.ClearConditions(o.operation)
		o.operation.Status.OperationID = oputils.NewOperationId()
	}
	if o.phaseIn(oputils.PhaseReconciling) {
		err := o.reconcilingApplications(ctx)
		if err != nil {
			o.logger.Error(err, "reconciling applications failed")
			o.recorder.Event(o.operation, "Warning", "ReconcileFailed", "Failed to reconcile deployments")
			return reconciler.RequeueWithError(err)
		}

		o.operation.Status.Phase = oputils.PhaseReconciled
		return reconciler.RequeueOnErrorOrStop(o.client.Status().Update(ctx, o.operation))
	}

	// check the diff between the expected and actual apps, set phase to reconciling and requeue if changes
	expectedCacheKey := ctlrutils.NewCacheKeyFromApplications(o.operation.Spec.Applications)
	if o.operation.Status.CacheKey != expectedCacheKey {
		o.operation.Status.CacheKey = expectedCacheKey
		o.operation.Status.Phase = oputils.PhaseReconciling
	}
	return reconciler.RequeueOnErrorOrContinue(o.client.Status().Update(ctx, o.operation))
}

func (o *OperationAdapter) EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("operation EnsureFinalizer")
	if o.operation.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(o.operation, oputils.FinalizerName) {
		controllerutil.AddFinalizer(o.operation, oputils.FinalizerName)
	}
	return reconciler.RequeueOnErrorOrContinue(o.client.Update(ctx, o.operation))
}

func (o *OperationAdapter) EnsureFinalizerRemoved(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("operation EnsureFinalizerDeleted")
	if !o.operation.ObjectMeta.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(o.operation, oputils.FinalizerName) {
		if o.phaseIn(oputils.PhaseDeleted) {
			o.logger.V(1).Info("All app deleted removing finalizer")
			controllerutil.RemoveFinalizer(o.operation, oputils.FinalizerName)
			return reconciler.RequeueOnErrorOrContinue(o.client.Update(ctx, o.operation))
		}
		if !o.phaseIn(oputils.PhaseDeleting) {
			o.logger.V(1).Info("App is not deleted yet, setting phase to deleting")
			o.operation.Status.Phase = oputils.PhaseDeleting
			return reconciler.RequeueOnErrorOrContinue(o.client.Status().Update(ctx, o.operation))
		}
	}
	return reconciler.ContinueProcessing()
}

func (o *OperationAdapter) EnsureAllAppsAreDeleted(ctx context.Context) (reconciler.OperationResult, error) {
	o.logger.V(1).Info("Operation EnsureAllAppsAreDeleted")
	if o.phaseIn(oputils.PhaseDeleting) {
		// deleting logic here
		o.operation.Status.Phase = oputils.PhaseDeleted
		return reconciler.RequeueOnErrorOrStop(o.client.Status().Update(ctx, o.operation))
	}
	return reconciler.ContinueProcessing()
}

func (o *OperationAdapter) reconcilingApplications(ctx context.Context) error {
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

	added, removed, updated := oputils.DiffAppDeployments(expectedAppDeployments, currentAppDeployments, oputils.CompareProvisionJobs)
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
		if appdeployment.Status.Phase != apdutils.PhaseReady {
			return fmt.Errorf("app deployment is not ready: name %s, status, %s", app.Name, app.Status.Phase)
		}
	}

	return nil
}

func (o *OperationAdapter) expectedAppDeployments() []v1alpha1.AppDeployment {
	return lo.Map(o.operation.Spec.Applications, func(app v1alpha1.ApplicationSpec, index int) v1alpha1.AppDeployment {
		return v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      apdutils.OperationScopedAppDeployment(app.Name, o.operation.Status.OperationID),
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

func (o *OperationAdapter) listCurrentAppDeployments(ctx context.Context) ([]v1alpha1.AppDeployment, error) {
	appDeploymentList := &v1alpha1.AppDeploymentList{}
	if err := o.client.List(ctx, appDeploymentList, client.MatchingFields{operationOwnerKey: o.operation.Name}); err != nil {
		return nil, fmt.Errorf("failed to list appDeployments: %w", err)
	}
	return lo.Map(appDeploymentList.Items, func(app v1alpha1.AppDeployment, index int) v1alpha1.AppDeployment {
		return v1alpha1.AppDeployment{
			ObjectMeta: app.ObjectMeta,
			Spec:       app.Spec,
		}
	}), nil
}
