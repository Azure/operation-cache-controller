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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/internal/handler"
	hmocks "github.com/Azure/operation-cache-controller/internal/handler/mocks"
	utilsmock "github.com/Azure/operation-cache-controller/internal/utils/mocks"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

func newTestJobSpec() batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "test-image",
						Command: []string{
							"echo",
							"hello",
						},
						Args: []string{
							"world",
						},
					},
				},
			},
		},
	}
}

var _ = Describe("AppDeployment Controller", func() {
	Context("When setupWithManager is called", func() {
		It("Should setup the controller with the manager", func() {

			// Create a new mock controller
			mockCtrl := gomock.NewController(GinkgoT())
			defer mockCtrl.Finish()

			k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
				Scheme: scheme.Scheme,
			})
			Expect(err).NotTo(HaveOccurred())

			err = (&AppDeploymentReconciler{
				Client:   k8sManager.GetClient(),
				Scheme:   k8sManager.GetScheme(),
				recorder: k8sManager.GetEventRecorderFor("appdeployment-controller"),
			}).SetupWithManager(k8sManager)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		var (
			mockRecorderCtrl *gomock.Controller
			mockRecorder     *utilsmock.MockEventRecorder
			mockAdapterCtrl  *gomock.Controller
			mockAdapter      *hmocks.MockAppDeploymentHandlerInterface
		)
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		appdeployment := &v1alpha1.AppDeployment{
			Spec: v1alpha1.AppDeploymentSpec{
				OpId:      "test-op-id",
				Provision: newTestJobSpec(),
				Teardown:  newTestJobSpec(),
			},
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind AppDeployment")
			err := k8sClient.Get(ctx, typeNamespacedName, appdeployment)
			if err != nil && errors.IsNotFound(err) {
				resource := &v1alpha1.AppDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: v1alpha1.AppDeploymentSpec{
						Provision: newTestJobSpec(),
						Teardown:  newTestJobSpec(),
					}}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			mockRecorderCtrl = gomock.NewController(GinkgoT())
			mockRecorder = utilsmock.NewMockEventRecorder(mockRecorderCtrl)
			mockAdapterCtrl = gomock.NewController(GinkgoT())
			mockAdapter = hmocks.NewMockAppDeploymentHandlerInterface(mockAdapterCtrl)
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &v1alpha1.AppDeployment{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance AppDeployment")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AppDeploymentReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				recorder: mockRecorder,
			}
			ctx = context.WithValue(ctx, handler.AppdeploymentHandlerContextKey{}, mockAdapter)

			mockAdapter.EXPECT().EnsureApplicationValid(gomock.Any()).Return(reconciler.OperationResult{}, nil)
			mockAdapter.EXPECT().EnsureFinalizer(gomock.Any()).Return(reconciler.OperationResult{}, nil)
			mockAdapter.EXPECT().EnsureFinalizerDeleted(gomock.Any()).Return(reconciler.OperationResult{}, nil)
			mockAdapter.EXPECT().EnsureDependenciesReady(gomock.Any()).Return(reconciler.OperationResult{}, nil)
			mockAdapter.EXPECT().EnsureDeployingFinished(gomock.Any()).Return(reconciler.OperationResult{}, nil)
			mockAdapter.EXPECT().EnsureTeardownFinished(gomock.Any()).Return(reconciler.OperationResult{}, nil)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
		It("should cancel the reconcile loop", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AppDeploymentReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				recorder: mockRecorder,
			}
			ctx = context.WithValue(ctx, handler.AppdeploymentHandlerContextKey{}, mockAdapter)

			mockAdapter.EXPECT().EnsureApplicationValid(gomock.Any()).Return(reconciler.OperationResult{
				CancelRequest: true,
			}, nil)

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail to reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AppDeploymentReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				recorder: mockRecorder,
			}
			ctx = context.WithValue(ctx, handler.AppdeploymentHandlerContextKey{}, mockAdapter)

			mockAdapter.EXPECT().EnsureApplicationValid(gomock.Any()).Return(reconciler.OperationResult{}, errors.NewServiceUnavailable("test error"))

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(errors.IsServiceUnavailable(err)).To(BeTrue(), "expected error is ServiceUnavailable")
		})
	})

	Context("appDeploymentIndexerFunc tests", func() {
		It("should return nil for Job without owner", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "job-without-owner",
					Namespace: "default",
				},
			}

			result := appDeploymentIndexerFunc(job)
			Expect(result).To(BeNil())
		})

		It("should return nil for Job with non-AppDeployment owner", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "job-with-wrong-owner",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind:       "Pod",
							Name:       "owner-pod",
							UID:        "12345",
							Controller: &[]bool{true}[0],
						},
					},
				},
			}

			result := appDeploymentIndexerFunc(job)
			Expect(result).To(BeNil())
		})

		It("should return owner name for Job with AppDeployment owner", func() {
			ownerName := "test-appdeployment"
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "job-with-appdeployment-owner",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: v1alpha1.GroupVersion.String(),
							Kind:       "AppDeployment",
							Name:       ownerName,
							UID:        "67890",
							Controller: &[]bool{true}[0],
						},
					},
				},
			}

			result := appDeploymentIndexerFunc(job)
			Expect(result).To(Equal([]string{ownerName}))
		})

		It("should return nil for Job with AppDeployment owner reference that is not controller", func() {
			job := &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "job-with-non-controller-appdeployment",
					Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: v1alpha1.GroupVersion.String(),
							Kind:       "AppDeployment",
							Name:       "test-appdeployment",
							UID:        "13579",
							Controller: nil, // Not a controller reference
						},
					},
				},
			}

			result := appDeploymentIndexerFunc(job)
			Expect(result).To(BeNil())
		})
	})
})
