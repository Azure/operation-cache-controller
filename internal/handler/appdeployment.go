package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	apierror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/internal/log"
	ctrlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

type AppdeploymentHandlerContextKey struct{}

var (
	ErrJobFailed = errors.New("job failed")
)

//go:generate mockgen -destination=./mocks/mock_appdeployment.go -package=mocks github.com/Azure/operation-cache-controller/internal/handler AppDeploymentHandlerInterface
type AppDeploymentHandlerInterface interface {
	EnsureApplicationValid(ctx context.Context) (reconciler.OperationResult, error)
	EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error)
	EnsureFinalizerDeleted(ctx context.Context) (reconciler.OperationResult, error)
	EnsureDependenciesReady(ctx context.Context) (reconciler.OperationResult, error)
	EnsureDeployingFinished(ctx context.Context) (reconciler.OperationResult, error)
	EnsureTeardownFinished(ctx context.Context) (reconciler.OperationResult, error)
}

type AppDeploymentHandler struct {
	appDeployment *v1alpha1.AppDeployment
	logger        logr.Logger
	client        client.Client
	recorder      record.EventRecorder

	apdutil ctrlutils.AppDeploymentHelper
}

func NewAppDeploymentHandler(ctx context.Context, appDeployment *v1alpha1.AppDeployment, logger logr.Logger, client client.Client, recorder record.EventRecorder) AppDeploymentHandlerInterface {
	if appdeploymentHandler, ok := ctx.Value(AppdeploymentHandlerContextKey{}).(AppDeploymentHandlerInterface); ok {
		return appdeploymentHandler
	}
	return &AppDeploymentHandler{
		appDeployment: appDeployment,
		logger:        logger,
		recorder:      recorder,
		client:        client,

		apdutil: ctrlutils.NewAppDeploymentHelper(),
	}
}

func (a *AppDeploymentHandler) phaseIs(phase ...string) bool {
	for _, p := range phase {
		if a.appDeployment.Status.Phase == p {
			return true
		}
	}
	return false
}

func (a *AppDeploymentHandler) EnsureApplicationValid(ctx context.Context) (reconciler.OperationResult, error) {
	a.logger.V(1).Info("Operation EnsureApplicationValid")
	if err := ctrlutils.Validate(a.appDeployment); err != nil {
		a.recorder.Event(a.appDeployment, "Error", "InvalidApplication", err.Error())
		return reconciler.RequeueWithError(err)
	}
	// initialize the appdeployment status
	if a.phaseIs(v1alpha1.AppDeploymentPhaseEmpty) {
		a.logger.V(1).Info("Initializing appdeployment status")
		a.appDeployment.Status.Phase = v1alpha1.AppDeploymentPhasePending
		a.apdutil.ClearConditions(ctx, a.appDeployment)
		return reconciler.RequeueOnErrorOrContinue(a.client.Status().Update(ctx, a.appDeployment))
	}

	return reconciler.ContinueProcessing()
}

func (a *AppDeploymentHandler) EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error) {
	a.logger.V(1).Info("Operation EnsureFinalizer")
	if a.appDeployment.ObjectMeta.DeletionTimestamp.IsZero() && !controllerutil.ContainsFinalizer(a.appDeployment, v1alpha1.AppDeploymentFinalizerName) {
		controllerutil.AddFinalizer(a.appDeployment, v1alpha1.AppDeploymentFinalizerName)
	}
	return reconciler.RequeueOnErrorOrContinue(a.client.Update(ctx, a.appDeployment))
}

func (a *AppDeploymentHandler) EnsureFinalizerDeleted(ctx context.Context) (reconciler.OperationResult, error) {
	a.logger.V(1).Info("Operation EnsureFinalizerDeleted")
	if !a.appDeployment.ObjectMeta.DeletionTimestamp.IsZero() && controllerutil.ContainsFinalizer(a.appDeployment, v1alpha1.AppDeploymentFinalizerName) {
		if a.phaseIs(v1alpha1.AppDeploymentPhaseDeleted) {
			a.logger.V(1).Info("All app deleted removing finalizer")
			controllerutil.RemoveFinalizer(a.appDeployment, v1alpha1.AppDeploymentFinalizerName)
			return reconciler.RequeueOnErrorOrContinue(a.client.Update(ctx, a.appDeployment))
		}
		if !a.phaseIs(v1alpha1.AppDeploymentPhaseDeleting) {
			a.logger.V(1).Info("App is not deleted yet, setting phase to deleting")
			a.appDeployment.Status.Phase = v1alpha1.AppDeploymentPhaseDeleting
			return reconciler.RequeueOnErrorOrContinue(a.client.Status().Update(ctx, a.appDeployment))
		}
	}
	return reconciler.ContinueProcessing()
}

func (a *AppDeploymentHandler) EnsureDependenciesReady(ctx context.Context) (reconciler.OperationResult, error) {
	if !a.phaseIs(v1alpha1.AppDeploymentPhasePending) {
		return reconciler.ContinueProcessing()
	}
	a.logger.V(1).Info("Operation EnsureDependenciesReady")
	// list all dependencies and check if they are ready
	for _, dep := range a.appDeployment.Spec.Dependencies {
		// check if dependency is ready
		appdeployment := &v1alpha1.AppDeployment{}
		realAppName := ctrlutils.OperationScopedAppDeployment(dep, a.appDeployment.Spec.OpId)
		if err := a.client.Get(ctx, client.ObjectKey{Namespace: a.appDeployment.Namespace, Name: realAppName}, appdeployment); err != nil {
			a.logger.V(1).Error(err, "dependency not found", "dependency", realAppName)
			return reconciler.RequeueWithError(fmt.Errorf("dependency not found: %s ", realAppName))
		}
		if appdeployment.Status.Phase != v1alpha1.AppDeploymentPhaseReady {
			return reconciler.RequeueWithError(fmt.Errorf("dependency is not ready: %s", realAppName))
		}
	}
	// all dependencies are ready
	a.appDeployment.Status.Phase = v1alpha1.AppDeploymentPhaseDeploying
	return reconciler.RequeueOnErrorOrContinue(a.client.Status().Update(ctx, a.appDeployment))
}

var (
	errJobNotCompleted = fmt.Errorf("job not completed")
)

func (a *AppDeploymentHandler) createJob(ctx context.Context, jobTemplate *batchv1.Job) error {
	if err := ctrl.SetControllerReference(a.appDeployment, jobTemplate, a.client.Scheme()); err != nil {
		return fmt.Errorf("failed to set controller reference for job %s: %w", jobTemplate.Name, err)
	}
	if err := a.client.Create(ctx, jobTemplate); err != nil {
		return fmt.Errorf("failed to create job %s: %w", jobTemplate.Name, err)
	}
	return nil
}

func (a *AppDeploymentHandler) initializeJobAndAwaitCompletion(ctx context.Context, jobTemplate *batchv1.Job) error {
	job := &batchv1.Job{}
	// check if the job exists
	if err := a.client.Get(ctx, client.ObjectKey{Namespace: a.appDeployment.Namespace, Name: jobTemplate.Name}, job); err != nil {
		if !apierror.IsNotFound(err) {
			return fmt.Errorf("failed to get job %s: %w", jobTemplate.Name, err)
		}
		// create a new job
		if err := a.createJob(ctx, jobTemplate); err != nil {
			a.recorder.Event(a.appDeployment, "Error", "FailedCreateJob", err.Error())
			return fmt.Errorf("failed to create job %s: %w", jobTemplate.Name, err)
		}
		return errJobNotCompleted // requeue
	}

	// check if the job is running
	switch ctrlutils.CheckJobStatus(ctx, job) {
	// if job is failed then delete the job and create a new one
	case ctrlutils.JobStatusFailed:
		// delete the failed job
		if err := a.client.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
			a.recorder.Event(a.appDeployment, "Error", "FailedDeleteJob", err.Error())
			return fmt.Errorf("failed to delete job %s: %w", job.Name, err)
		}
		// complete the job if it is a teardown job
		if strings.HasPrefix(jobTemplate.Name, ctrlutils.JobTypeTeardown) {
			a.logger.Error(ErrJobFailed, "teardown job failed", log.FieldKeyAppDeploymentJobName, jobTemplate.Name)
			a.recorder.Event(a.appDeployment, "Warning", "TeardownJobFailed", fmt.Sprintf("Teardown job %s failed, requeuing for retry", jobTemplate.Name))
			// return nil to make the teardown job complete
			return nil
		}

		// create a new job if it is not a teardown job
		if err := ctrl.SetControllerReference(a.appDeployment, jobTemplate, a.client.Scheme()); err != nil {
			return fmt.Errorf("failed to set controller reference for job %s: %w", job.Name, err)
		}
		if err := a.client.Create(ctx, jobTemplate); err != nil {
			return fmt.Errorf("failed to create job %s: %w", jobTemplate.Name, err)
		}

	// if job is succeeded then delete the job
	case ctrlutils.JobStatusSucceeded:
		// delete the succeeded job
		if err := a.client.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete succeeded job %s: %w", job.Name, err)
		}
		return nil
	}
	return errJobNotCompleted
}

// EnsureDeployingFinished checks if the provision job exists
// if not exist then create a new provision job
// if job is exist && running then requeue and waiting for the job complete
// if job is exist && failed then delete the job and create a new one
// if job is exist && succeeded then update the appdeployment status to ready
func (a *AppDeploymentHandler) EnsureDeployingFinished(ctx context.Context) (reconciler.OperationResult, error) {
	a.logger.V(1).Info("Operation EnsureDeployingFinished")
	if !a.phaseIs(v1alpha1.AppDeploymentPhaseDeploying) {
		return reconciler.ContinueProcessing()
	}
	provisionJob := ctrlutils.ProvisionJobFromAppDeploymentSpec(a.appDeployment)
	err := a.initializeJobAndAwaitCompletion(ctx, provisionJob)
	switch err {
	case nil:
		// provision job is succeeded move the appdeployment to ready phase
		a.appDeployment.Status.Phase = v1alpha1.AppDeploymentPhaseReady
		return reconciler.RequeueOnErrorOrContinue(a.client.Status().Update(ctx, a.appDeployment))
	case errJobNotCompleted:
		a.logger.V(1).WithValues(log.FieldKeyAppDeploymentJobName, provisionJob.Name).Info("provision job is not completed yet")
		return reconciler.Requeue()
	default:
		a.logger.Error(err, "provision job failed %s", provisionJob.Name)
		return reconciler.RequeueWithError(err)
	}
}

func (a *AppDeploymentHandler) EnsureTeardownFinished(ctx context.Context) (reconciler.OperationResult, error) {
	a.logger.V(1).Info("Operation EnsureTeardownFinished")
	if !a.phaseIs(v1alpha1.AppDeploymentPhaseDeleting) {
		return reconciler.ContinueProcessing()
	}
	teardownJob := ctrlutils.TeardownJobFromAppDeploymentSpec(a.appDeployment)
	err := a.initializeJobAndAwaitCompletion(ctx, teardownJob)
	switch err {
	case nil:
		// teardown job is succeeded move the appdeployment to deleted phase
		a.appDeployment.Status.Phase = v1alpha1.AppDeploymentPhaseDeleted
		return reconciler.RequeueOnErrorOrContinue(a.client.Status().Update(ctx, a.appDeployment))
	case errJobNotCompleted:
		a.logger.V(1).WithValues(log.FieldKeyAppDeploymentJobName, teardownJob.Name).Info("teardown job is not completed yet")
		return reconciler.Requeue()
	default:
		a.logger.WithValues(log.FieldKeyAppDeploymentJobName, teardownJob.Name).Error(err, "teardown job failed %s")
		return reconciler.RequeueWithError(err)
	}
}
