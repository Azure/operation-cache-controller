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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/internal/handler"
	"github.com/Azure/operation-cache-controller/internal/handler/mocks"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

var _ = Describe("Requirement Controller", func() {
	Context("When setupWithManager is called", func() {
		It("Should setup the controller with the manager", func() {

			// Create a new mock controller
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme.Scheme,
			})
			Expect(err).NotTo(HaveOccurred())

			err = (&RequirementReconciler{
				Client:   k8sManager.GetClient(),
				Scheme:   k8sManager.GetScheme(),
				recorder: k8sManager.GetEventRecorderFor("requirement-controller"),
			}).SetupWithManager(k8sManager)

			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("When creating a new Requirement Controller", func() {
		var (
			timeout  = time.Second * 10
			interval = time.Millisecond * 250
		)
		It("Should create a new Requirement Controller", func() {
			key := types.NamespacedName{
				Name:      "test-requirement",
				Namespace: "default",
			}
			requirement := &v1alpha1.Requirement{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.RequirementSpec{
					Template: v1alpha1.OperationSpec{
						Applications: []v1alpha1.ApplicationSpec{
							{
								Name: "test-app",
								Provision: batchv1.JobSpec{
									Template: corev1.PodTemplateSpec{
										Spec: corev1.PodSpec{
											Containers: []corev1.Container{
												{
													Image: "test-image",
												},
											},
										},
									},
								},
								Teardown: batchv1.JobSpec{
									Template: corev1.PodTemplateSpec{
										Spec: corev1.PodSpec{
											Containers: []corev1.Container{
												{
													Image: "test-image",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), requirement)).Should(Succeed())

			fetchedRequirement := &v1alpha1.Requirement{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), key, fetchedRequirement)
				return err == nil
			}, timeout, interval).Should(BeTrue())
		})
	})
	Context("When reconciling a resource with adapter", func() {
		var (
			mockClientCtrl   *gomock.Controller
			mockRecorderCtrl *gomock.Controller
			mockAdapterCtrl  *gomock.Controller
			mockAdapter      *mocks.MockRequirementHandlerInterface

			requirementReconciler *RequirementReconciler

			key = types.NamespacedName{
				Name:      "test-requirement",
				Namespace: "default",
			}
		)

		BeforeEach(func() {
			mockClientCtrl = gomock.NewController(GinkgoT())
			mockRecorderCtrl = gomock.NewController(GinkgoT())
			mockAdapterCtrl = gomock.NewController(GinkgoT())
			mockAdapter = mocks.NewMockRequirementHandlerInterface(mockAdapterCtrl)

			requirementReconciler = &RequirementReconciler{
				Client: k8sClient,
				Scheme: scheme.Scheme,
			}
		})

		AfterEach(func() {
			mockClientCtrl.Finish()
			mockRecorderCtrl.Finish()
			mockAdapterCtrl.Finish()
		})

		It("Should reconcile the resource with adapter", func() {
			mockAdapter.EXPECT().EnsureNotExpired(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureInitialized(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureCacheExisted(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureCachedOperationAcquired(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)
			mockAdapter.EXPECT().EnsureOperationReady(gomock.Any()).Return(reconciler.ContinueOperationResult(), nil)

			result, err := requirementReconciler.Reconcile(context.WithValue(context.Background(), handler.RequiremenContextKey{}, mockAdapter), ctrl.Request{
				NamespacedName: key,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{RequeueAfter: defaultCheckInterval}))
		})
		It("Should return error if any", func() {
			By("Reconciling the resource with adapter")
			ctx := context.Background()
			testErr := fmt.Errorf("test-error")
			mockAdapter.EXPECT().EnsureNotExpired(gomock.Any()).Return(reconciler.RequeueWithError(testErr))

			result, err := requirementReconciler.Reconcile(context.WithValue(ctx, handler.RequiremenContextKey{}, mockAdapter), ctrl.Request{
				NamespacedName: key,
			})
			Expect(err).To(MatchError(testErr))
			Expect(result).Should(Equal(reconcile.Result{}))
		})

		It("should cancel the reconcile loop", func() {
			By("Reconciling the created resource")
			ctx := context.Background()

			mockAdapter.EXPECT().EnsureNotExpired(gomock.Any()).Return(reconciler.OperationResult{
				CancelRequest: true,
			}, nil)

			result, err := requirementReconciler.Reconcile(context.WithValue(ctx, handler.RequiremenContextKey{}, mockAdapter), ctrl.Request{
				NamespacedName: key,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{}))
		})
	})
	Context("When reconciling a resource", func() {
		const resourceName = "test-req-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		requirement := &v1alpha1.Requirement{}

		BeforeEach(func() {
			By("Creating a new Requirement resource")
			err := k8sClient.Get(ctx, typeNamespacedName, requirement)
			if err != nil && errors.IsNotFound(err) {
				resource := &v1alpha1.Requirement{
					ObjectMeta: metav1.ObjectMeta{
						Name:      typeNamespacedName.Name,
						Namespace: typeNamespacedName.Namespace,
					},
					Spec: v1alpha1.RequirementSpec{
						Template: v1alpha1.OperationSpec{
							Applications: []v1alpha1.ApplicationSpec{
								{
									Name: "test-app",
									Provision: batchv1.JobSpec{
										Template: corev1.PodTemplateSpec{
											Spec: corev1.PodSpec{
												Containers: []corev1.Container{
													{
														Image: "test-image",
													},
												},
											},
										},
									},
									Teardown: batchv1.JobSpec{
										Template: corev1.PodTemplateSpec{
											Spec: corev1.PodSpec{
												Containers: []corev1.Container{
													{
														Image: "test-image",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(context.Background(), resource)).Should(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &v1alpha1.Requirement{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Requirement")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RequirementReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			result, err := controllerReconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).Should(Equal(reconcile.Result{RequeueAfter: reconciler.DefaultRequeueDelay}))
		})
	})
})
