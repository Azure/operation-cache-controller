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
	"github.com/Azure/operation-cache-controller/internal/handler"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

const (
	defaultCacheCheckInterval = time.Second * 60
)

// CacheReconciler reconciles a Cache object
type CacheReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=controller.azure.github.com,resources=caches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=caches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=caches/finalizers,verbs=update
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=operations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=controller.azure.github.com,resources=operations/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cache object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *CacheReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	cache := &v1alpha1.Cache{}
	if err := r.Get(ctx, req.NamespacedName, cache); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return r.reconcileHandler(ctx, handler.NewCacheHandler(ctx, cache, logger, r.Client, r.Scheme, r.recorder, ctrl.SetControllerReference))
}

func (r *CacheReconciler) reconcileHandler(ctx context.Context, h handler.CacheHandlerInterface) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	operations := []reconciler.ReconcileOperation{
		h.CheckCacheExpiry,
		h.EnsureCacheInitialized,
		h.CalculateKeepAliveCount,
		h.AdjustCache,
	}

	for _, operation := range operations {
		result, err := operation(ctx)
		if err != nil || result.RequeueRequest {
			logger.Error(err, "cache operation failed")
			return ctrl.Result{RequeueAfter: result.RequeueDelay}, err
		}
		if result.CancelRequest {
			logger.Info("cache reconcile canceled, requeue after 60 seconds")
			return ctrl.Result{RequeueAfter: defaultCacheCheckInterval}, nil
		}
	}
	logger.Info("cache reconcile completed, requeue after 60 seconds")
	return ctrl.Result{RequeueAfter: defaultCacheCheckInterval}, nil
}

func cacheOperationIndexerFunc(obj client.Object) []string {
	// grab the operation object, extract the owner...
	operation := obj.(*v1alpha1.Operation)
	owner := metav1.GetControllerOf(operation)
	if owner == nil {
		return nil
	}
	// ...make sure it's a Cache...
	if owner.APIVersion != v1alpha1.GroupVersion.String() || owner.Kind != "Cache" {
		return nil
	}
	// ...and if so, return it
	return []string{owner.Name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *CacheReconciler) SetupWithManager(mgr ctrl.Manager) error { // +gocover:ignore:block init controller
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &v1alpha1.Operation{}, v1alpha1.CacheOwnerKey, cacheOperationIndexerFunc); err != nil { // +gocover:ignore:block init controller
		return err
	}
	// +gocover:ignore:block init controller
	r.recorder = mgr.GetEventRecorderFor("Cache")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Cache{}).
		Owns(&v1alpha1.Operation{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 50,
		}).
		Named("cache").
		Complete(r)
}
