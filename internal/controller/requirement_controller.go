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
	"time"

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

var defaultCheckInterval = 10 * time.Minute

// RequirementReconciler reconciles a Requirement object
type RequirementReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=controller.azure.github.com,resources=requirements,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=requirements/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=requirements/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Requirement object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *RequirementReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("requirement", req.NamespacedName)

	requirement := &v1alpha1.Requirement{}
	if err := r.Get(ctx, req.NamespacedName, requirement); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	adapter := NewRequirementAdapter(ctx, requirement, logger, r.Client, r.recorder)
	return r.ReconcileHandler(ctx, adapter)
}
func (r *RequirementReconciler) ReconcileHandler(ctx context.Context, adapter RequirementAdapterInterface) (ctrl.Result, error) {
	operations := []reconciler.ReconcileOperation{
		adapter.EnsureNotExpired,
		adapter.EnsureInitialized,
		adapter.EnsureCacheExisted,
		adapter.EnsureCachedOperationAcquired,
		adapter.EnsureOperationReady,
	}

	for _, operation := range operations {
		result, err := operation(ctx)
		if err != nil {
			return ctrl.Result{}, err
		}
		if result.RequeueRequest {
			return ctrl.Result{RequeueAfter: reconciler.DefaultRequeueDelay}, err
		}
		if result.CancelRequest {
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{RequeueAfter: defaultCheckInterval}, nil
}

var requirementOwnerKey = ".requirement.metadata.controller"

// SetupWithManager sets up the controller with the Manager.
func (r *RequirementReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.Operation{}, requirementOwnerKey,
		func(rawObj client.Object) []string {
			// grab the Operation object, extract the owner
			op := rawObj.(*v1alpha1.Operation)
			owner := metav1.GetControllerOf(op)
			if owner == nil {
				return nil
			}
			// Make sure the owner is a Requirement object
			if owner.APIVersion != v1alpha1.GroupVersion.String() || owner.Kind != "Requirement" {
				return nil
			}
			return []string{owner.Name}
		}); err != nil {
		return err
	}

	r.recorder = mgr.GetEventRecorderFor("Requirement")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Requirement{}).
		Owns(&v1alpha1.Operation{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 100,
		}).
		Named("requirement").
		Complete(r)
}
