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
	"testing"

	// "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/Azure/operation-cache-controller/internal/controller/mocks"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
)

func TestReconcile(t *testing.T) {
	ctx := context.Background()

	t.Run("get cache failed", func(t *testing.T) {
		builder := fake.NewClientBuilder()
		buildReconciler := CacheReconciler{
			Client: builder.Build(),
		}
		_, err := buildReconciler.Reconcile(ctx, ctrl.Request{})
		assert.Error(t, err)
	})
}

func TestReconcileHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("reconcile successfully", func(t *testing.T) {
		builder := fake.NewClientBuilder()
		scheme := runtime.NewScheme()

		builder.WithScheme(scheme)

		cacheReconciler := CacheReconciler{
			Client: builder.Build(),
			Scheme: scheme,
		}
		mockCacheAdapterCtrl := gomock.NewController(t)
		cacheAdapter := mocks.NewMockCacheAdapterInterface(mockCacheAdapterCtrl)
		cacheAdapter.EXPECT().CheckCacheExpiry(ctx).Return(reconciler.OperationResult{}, nil)
		cacheAdapter.EXPECT().EnsureCacheInitialized(ctx).Return(reconciler.OperationResult{}, nil)
		cacheAdapter.EXPECT().CalculateKeepAliveCount(ctx).Return(reconciler.OperationResult{}, nil)
		cacheAdapter.EXPECT().AdjustCache(ctx).Return(reconciler.OperationResult{}, nil)
		res, err := cacheReconciler.reconcileHandler(ctx, cacheAdapter)
		assert.NoError(t, err)
		assert.Equal(t, defaultCacheCheckInterval, res.RequeueAfter)
	})

	t.Run("reconcile canceled", func(t *testing.T) {
		builder := fake.NewClientBuilder()
		scheme := runtime.NewScheme()

		builder.WithScheme(scheme)

		cacheReconciler := CacheReconciler{
			Client: builder.Build(),
			Scheme: scheme,
		}
		mockCacheAdapterCtrl := gomock.NewController(t)
		cacheAdapter := mocks.NewMockCacheAdapterInterface(mockCacheAdapterCtrl)
		cacheAdapter.EXPECT().CheckCacheExpiry(ctx).Return(reconciler.OperationResult{CancelRequest: true}, nil)
		res, err := cacheReconciler.reconcileHandler(ctx, cacheAdapter)
		assert.NoError(t, err)
		assert.Equal(t, defaultCacheCheckInterval, res.RequeueAfter)
	})

	t.Run("reconcile err", func(t *testing.T) {
		builder := fake.NewClientBuilder()
		scheme := runtime.NewScheme()

		builder.WithScheme(scheme)

		cacheReconciler := CacheReconciler{
			Client: builder.Build(),
			Scheme: scheme,
		}
		mockCacheAdapterCtrl := gomock.NewController(t)
		cacheAdapter := mocks.NewMockCacheAdapterInterface(mockCacheAdapterCtrl)
		cacheAdapter.EXPECT().CheckCacheExpiry(ctx).Return(reconciler.OperationResult{}, assert.AnError)
		_, err := cacheReconciler.reconcileHandler(ctx, cacheAdapter)
		assert.NotNil(t, err)
	})
}

func TestCacheOperationIndexerFunc(t *testing.T) {
	t.Run("with Cache owner", func(t *testing.T) {
		// Create an operation with a Cache owner
		operation := &appsv1.Operation{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-operation",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: appsv1.GroupVersion.String(),
						Kind:       "Cache",
						Name:       "test-cache",
						Controller: func() *bool { b := true; return &b }(),
					},
				},
			},
		}

		// Call the indexer function
		result := cacheOperationIndexerFunc(operation)

		// Verify the result
		assert.Equal(t, []string{"test-cache"}, result)
	})

	t.Run("with no owner", func(t *testing.T) {
		// Create an operation with no owner
		operation := &appsv1.Operation{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-operation",
			},
		}

		// Call the indexer function
		result := cacheOperationIndexerFunc(operation)

		// Verify the result
		assert.Nil(t, result)
	})

	t.Run("with non-Cache owner", func(t *testing.T) {
		// Create an operation with a non-Cache owner
		operation := &appsv1.Operation{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-operation",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: appsv1.GroupVersion.String(),
						Kind:       "Requirement",
						Name:       "test-requirement",
						Controller: func() *bool { b := true; return &b }(),
					},
				},
			},
		}

		// Call the indexer function
		result := cacheOperationIndexerFunc(operation)

		// Verify the result
		assert.Nil(t, result)
	})

	t.Run("with non-controller owner", func(t *testing.T) {
		// Create an operation with a non-controller owner reference
		operation := &appsv1.Operation{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-operation",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: appsv1.GroupVersion.String(),
						Kind:       "Cache",
						Name:       "test-cache",
						Controller: func() *bool { b := false; return &b }(),
					},
				},
			},
		}

		// Call the indexer function
		result := cacheOperationIndexerFunc(operation)

		// Verify the result
		assert.Nil(t, result)
	})

	t.Run("with different API version", func(t *testing.T) {
		// Create an operation with a different API version
		operation := &appsv1.Operation{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-operation",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "different.group/v1",
						Kind:       "Cache",
						Name:       "test-cache",
						Controller: func() *bool { b := true; return &b }(),
					},
				},
			},
		}

		// Call the indexer function
		result := cacheOperationIndexerFunc(operation)

		// Verify the result
		assert.Nil(t, result)
	})
}

var _ = Describe("Cache Controller", func() {
	Context("When setupWithManager is called", func() {
		It("should set up the controller with the manager", func() {

			// Create a new mock controller
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme.Scheme,
			})
			Expect(err).NotTo(HaveOccurred())

			err = (&CacheReconciler{
				Client:   k8sManager.GetClient(),
				Scheme:   k8sManager.GetScheme(),
				recorder: k8sManager.GetEventRecorderFor("appdeployment-controller"),
			}).SetupWithManager(k8sManager)

			Expect(err).NotTo(HaveOccurred())
		})
	})
	// TODO: need to figure out how to test the controller with indexer
	// Context("When reconciling a resource", func() {
	// 	const resourceName = "test-resource"

	// 	ctx := context.Background()

	// 	typeNamespacedName := types.NamespacedName{
	// 		Name:      resourceName,
	// 		Namespace: "default", // TODO(user):Modify as needed
	// 	}
	// 	cache := &appv1.Cache{}

	// 	BeforeEach(func() {
	// 		By("creating the custom resource for the Kind Cache")
	// 		err := k8sClient.Get(ctx, typeNamespacedName, cache)
	// 		if err != nil && errors.IsNotFound(err) {
	// 			resource := &appv1.Cache{
	// 				ObjectMeta: metav1.ObjectMeta{
	// 					Name:      resourceName,
	// 					Namespace: "default",
	// 				},
	// 				Spec: appv1.CacheSpec{
	// 					OperationTemplate: appv1.OperationSpec{
	// 						Applications: []appv1.ApplicationSpec{
	// 							{
	// 								Name:      "app1",
	// 								Provision: newTestJobSpec(),
	// 								Teardown:  newTestJobSpec(),
	// 							},
	// 						},
	// 					},
	// 					// Format time in UTC with Z suffix
	// 					ExpireTime: time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02T15:04:05Z"),
	// 				},
	// 			}
	// 			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
	// 		}
	// 	})

	// 	AfterEach(func() {
	// 		// TODO(user): Cleanup logic after each test, like removing the resource instance.
	// 		resource := &appv1.Cache{}
	// 		err := k8sClient.Get(ctx, typeNamespacedName, resource)
	// 		Expect(err).NotTo(HaveOccurred())

	// 		By("Cleanup the specific resource instance Cache")
	// 		Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
	// 	})
	// 	It("should successfully reconcile the resource", func() {
	// 		By("Reconciling the created resource")
	// 		controllerReconciler := &CacheReconciler{
	// 			Client: k8sClient,
	// 			Scheme: k8sClient.Scheme(),
	// 		}

	// 		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
	// 			NamespacedName: typeNamespacedName,
	// 		})
	// 		Expect(err).NotTo(HaveOccurred())
	// 		// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
	// 		// Example: If you expect a certain status condition after reconciliation, verify it here.
	// 	})
	// })
})
