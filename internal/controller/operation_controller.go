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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

// OperationReconciler reconciles a Operation object
type OperationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=controller.azure.github.com,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=operations/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=operations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Operation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *OperationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	operation := &v1alpha1.Operation{}
	if err := r.Get(ctx, req.NamespacedName, operation); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	adapter := NewOperationAdapter(ctx, operation, logger, r.Client, r.recorder)
	return r.ReconcileHandler(ctx, adapter)
}

func (r *OperationReconciler) ReconcileHandler(ctx context.Context, adapter OperationAdapterInterface) (ctrl.Result, error) {
	operations := []reconciler.ReconcileOperation{
		adapter.EnsureFinalizer,
		adapter.EnsureFinalizerRemoved,
		adapter.EnsureNotExpired,
		adapter.EnsureAllAppsAreReady,
		adapter.EnsureAllAppsAreDeleted,
	}

	for _, operation := range operations {
		result, err := operation(ctx)
		if err != nil || result.RequeueRequest {
			return ctrl.Result{RequeueAfter: reconciler.DefaultRequeueDelay}, err
		}
		if result.CancelRequest {
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

var operationOwnerKey = ".operation.metadata.controller"

// SetupWithManager sets up the controller with the Manager.
func (r *OperationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.AppDeployment{}, operationOwnerKey,
		func(rawObj client.Object) []string {
			// grab the AppDeployment object, extract the owner
			adp := rawObj.(*v1alpha1.AppDeployment)
			owner := metav1.GetControllerOf(adp)
			if owner == nil {
				return nil
			}
			// Make sure the owner is a Operation object
			if owner.APIVersion != v1alpha1.GroupVersion.String() || owner.Kind != "Operation" {
				return nil
			}
			return []string{owner.Name}
		}); err != nil {
		return err
	}

	r.recorder = mgr.GetEventRecorderFor("Operation")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Operation{}).
		Owns(&v1alpha1.AppDeployment{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 100,
		}).
		Named("operation").
		Complete(r)
}
