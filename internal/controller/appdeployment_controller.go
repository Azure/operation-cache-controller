/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	apdutil "github.com/Azure/operation-cache-controller/internal/utils/controller/appdeployment"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

// AppDeploymentReconciler reconciles a AppDeployment object
type AppDeploymentReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=controller.azure.github.com,resources=appdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=appdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=appdeployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=jobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AppDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues(apdutil.LogKeyAppDeploymentName, req.NamespacedName)
	appdeployment := &v1alpha1.AppDeployment{}
	if err := r.Get(ctx, req.NamespacedName, appdeployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	adapter := NewAppDeploymentAdapter(ctx, appdeployment, logger, r.Client, r.recorder)
	return r.ReconcileHandler(ctx, adapter)
}
func (r *AppDeploymentReconciler) ReconcileHandler(ctx context.Context, adapter AppDeploymentAdapterInterface) (ctrl.Result, error) {
	operations := []reconciler.ReconcileOperation{
		adapter.EnsureApplicationValid,
		adapter.EnsureFinalizer,
		adapter.EnsureFinalizerDeleted,
		adapter.EnsureDependenciesReady,
		adapter.EnsureDeployingFinished,
		adapter.EnsureTeardownFinished,
	}

	for _, operation := range operations {
		operationResult, err := operation(ctx)
		if err != nil || operationResult.RequeueRequest {
			return ctrl.Result{RequeueAfter: operationResult.RequeueDelay}, err
		}
		if operationResult.CancelRequest {
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

var appDeploymentOwnerKey = ".appDeployment.metadata.controller"

// SetupWithManager sets up the controller with the Manager.
func (r *AppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &batchv1.Job{}, appDeploymentOwnerKey,
		func(rawObj client.Object) []string {
			job := rawObj.(*batchv1.Job)
			owner := metav1.GetControllerOf(job)
			if owner == nil {
				return nil
			}
			if owner.APIVersion != v1alpha1.GroupVersion.String() || owner.Kind != "AppDeployment" {
				return nil
			}
			return []string{owner.Name}
		}); err != nil {
		return err
	}

	r.recorder = mgr.GetEventRecorderFor("AppDeployment")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AppDeployment{}).
		Owns(&batchv1.Job{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 100,
		}).
		Named("appdeployment").
		Complete(r)
}
