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
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/internal/handler"
	"github.com/Azure/operation-cache-controller/internal/handler/mocks"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

var _ = Describe("Operation Controller", func() {
	Context("When setupWithManager is called", func() {
		It("Should setup the controller with the manager", func() {

			// Create a new mock controller
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme.Scheme,
			})
			Expect(err).NotTo(HaveOccurred())

			err = (&OperationReconciler{
				Client:   k8sManager.GetClient(),
				Scheme:   k8sManager.GetScheme(),
				recorder: k8sManager.GetEventRecorderFor("appdeployment-controller"),
			}).SetupWithManager(k8sManager)

			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("When creating a new Operation Controller", func() {
		var (
			timeout  = time.Second * 10
			interval = time.Millisecond * 250
		)
		It("Should create a new Operation Controller", func() {
			key := types.NamespacedName{
				Name:      "test-operation",
				Namespace: "default",
			}
			operation := &v1alpha1.Operation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.OperationSpec{
					Applications: []v1alpha1.ApplicationSpec{
						{
							Name:      "test-app1",
							Provision: newTestJobSpec(),
							Teardown:  newTestJobSpec(),
						},
						{
							Name:         "test-app2",
							Provision:    newTestJobSpec(),
							Teardown:     newTestJobSpec(),
							Dependencies: []string{"test-app1"},
						},
					},
				},
			}

			Expect(k8sClient.Create(context.Background(), operation)).To(Succeed())

			feched := &v1alpha1.Operation{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), key, feched)
				return err == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
	Context("When reconciling a resource with adapter", func() {
		var (
			mockClientCtrl   *gomock.Controller
			mockRecorderCtrl *gomock.Controller
			mockAdapterCtrl  *gomock.Controller
			mockAdapter      *mocks.MockOperationHandlerInterface

			operationReconciler *OperationReconciler

			key = types.NamespacedName{
				Name:      "test-operation",
				Namespace: "default",
			}
		)

		BeforeEach(func() {
			mockClientCtrl = gomock.NewController(GinkgoT())
			mockRecorderCtrl = gomock.NewController(GinkgoT())
			mockAdapterCtrl = gomock.NewController(GinkgoT())
			mockAdapter = mocks.NewMockOperationHandlerInterface(mockAdapterCtrl)
			operationReconciler = &OperationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		AfterEach(func() {
			mockClientCtrl.Finish()
			mockRecorderCtrl.Finish()
			mockAdapterCtrl.Finish()
		})

		It("Should reconcile the resource with adapter", func() {
			By("Reconciling the created resource")
			ctx := context.Background()

			mockAdapter.EXPECT().EnsureFinalizer(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureFinalizerRemoved(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureNotExpired(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureAllAppsAreReady(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureAllAppsAreDeleted(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)

			res, err := operationReconciler.Reconcile(context.WithValue(ctx, handler.OperationContextKey{}, mockAdapter), reconcile.Request{
				NamespacedName: key,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{}))

		})
		It("should return the error if any", func() {
			By("Reconciling the created resource")
			ctx := context.Background()
			testErr := fmt.Errorf("test-error")
			mockAdapter.EXPECT().EnsureFinalizer(gomock.Any()).Return(reconciler.ContinueOperationResult(), testErr)

			res, err := operationReconciler.Reconcile(context.WithValue(ctx, handler.OperationContextKey{}, mockAdapter), reconcile.Request{
				NamespacedName: key,
			})
			Expect(err).To(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{RequeueAfter: reconciler.DefaultRequeueDelay}))
		})

		It("should cancel the reconcile loop", func() {
			By("Reconciling the created resource")
			ctx := context.Background()

			mockAdapter.EXPECT().EnsureFinalizer(gomock.Any()).Return(reconciler.OperationResult{
				CancelRequest: true,
			}, nil)

			res, err := operationReconciler.Reconcile(context.WithValue(ctx, handler.OperationContextKey{}, mockAdapter), reconcile.Request{
				NamespacedName: key,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).Should(Equal(reconcile.Result{}))
		})
	})
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		operation := &v1alpha1.Operation{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Operation")
			err := k8sClient.Get(ctx, typeNamespacedName, operation)
			if err != nil && errors.IsNotFound(err) {
				resource := &v1alpha1.Operation{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: v1alpha1.OperationSpec{
						Applications: []v1alpha1.ApplicationSpec{
							{
								Name:      "test-app1",
								Provision: newTestJobSpec(),
								Teardown:  newTestJobSpec(),
							},
							{
								Name:         "test-app2",
								Provision:    newTestJobSpec(),
								Teardown:     newTestJobSpec(),
								Dependencies: []string{"test-app1"},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &v1alpha1.Operation{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Operation")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &OperationReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})

func TestOperationIndexerFunc(t *testing.T) {
	// Test case 1: AppDeployment without an owner reference
	t.Run("AppDeployment without an owner reference", func(t *testing.T) {
		appDeployment := &v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-without-owner",
				Namespace: "default",
			},
		}
		result := operationIndexerFunc(appDeployment)
		assert.Nil(t, result, "Expected nil for AppDeployment without owner reference")
	})

	// Test case 2: AppDeployment with a non-controller owner reference
	t.Run("AppDeployment with a non-controller owner reference", func(t *testing.T) {
		appDeployment := &v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-with-non-controller-owner",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: v1alpha1.GroupVersion.String(),
						Kind:       "Operation",
						Name:       "test-operation",
						UID:        "12345",
						Controller: nil, // Not a controller reference
					},
				},
			},
		}
		result := operationIndexerFunc(appDeployment)
		assert.Nil(t, result, "Expected nil for AppDeployment with non-controller owner reference")
	})

	// Test case 3: AppDeployment with an owner that's not an Operation
	t.Run("AppDeployment with an owner that's not an Operation", func(t *testing.T) {
		appDeployment := &v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-with-wrong-owner-kind",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: v1alpha1.GroupVersion.String(),
						Kind:       "SomeOtherKind",
						Name:       "test-owner",
						UID:        "12345",
						Controller: &[]bool{true}[0],
					},
				},
			},
		}
		result := operationIndexerFunc(appDeployment)
		assert.Nil(t, result, "Expected nil for AppDeployment with wrong owner kind")
	})

	// Test case 4: AppDeployment with a valid Operation controller reference
	t.Run("AppDeployment with a valid Operation controller reference", func(t *testing.T) {
		operationName := "test-operation"
		appDeployment := &v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-with-operation-owner",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: v1alpha1.GroupVersion.String(),
						Kind:       "Operation",
						Name:       operationName,
						UID:        "12345",
						Controller: &[]bool{true}[0],
					},
				},
			},
		}
		result := operationIndexerFunc(appDeployment)
		assert.Equal(t, []string{operationName}, result, "Expected owner name for AppDeployment with Operation controller reference")
	})

	// Test case 5: AppDeployment with wrong API version owner reference
	t.Run("AppDeployment with wrong API version owner reference", func(t *testing.T) {
		appDeployment := &v1alpha1.AppDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "app-with-wrong-api-version",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: "wrong.api/v1",
						Kind:       "Operation",
						Name:       "test-operation",
						UID:        "12345",
						Controller: &[]bool{true}[0],
					},
				},
			},
		}
		result := operationIndexerFunc(appDeployment)
		assert.Nil(t, result, "Expected nil for AppDeployment with wrong API version")
	})
}
