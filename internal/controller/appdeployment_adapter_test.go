package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	batchv1 "k8s.io/api/batch/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/Azure/operation-cache-controller/api/v1alpha1"
	mockpkg "github.com/Azure/operation-cache-controller/internal/mocks"
	apdutil "github.com/Azure/operation-cache-controller/internal/utils/controller/appdeployment"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

const testOpId = "test-op-id"

var validAppDeployment = &v1alpha1.AppDeployment{
	Spec: v1alpha1.AppDeploymentSpec{
		Provision: newTestJobSpec(),
		Teardown:  newTestJobSpec(),
		OpId:      testOpId,
	},
}

func TestNewAppDeploymentAdapter(t *testing.T) {
	ctx := context.Background()
	appDeployment := validAppDeployment.DeepCopy()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorder := mockpkg.NewMockEventRecorder(mockCtrl)

	adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
	assert.NotNil(t, adapter)
}

func TestAppDeploymentAdapter_EnsureApplicationValid(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockClient := mockpkg.NewMockClient(mockCtrl)

	mockRecorderCtrl := gomock.NewController(t)
	defer mockRecorderCtrl.Finish()
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

	mockStatusCtrl := gomock.NewController(t)
	defer mockStatusCtrl.Finish()
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)

	mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
	mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockRecorder.EXPECT().Event(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	t.Run("Happy path: application valid", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)
		res, err := adapter.EnsureApplicationValid(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})

	t.Run("Happy path: application invalid and not in empty phase", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeploying
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)
		res, err := adapter.EnsureApplicationValid(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("Sad path: application return error", func(t *testing.T) {
		appDeployment := &v1alpha1.AppDeployment{}
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)
		res, err := adapter.EnsureApplicationValid(ctx)
		assert.Error(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}

func TestAppDeploymentAdapter_EnsureFinalizer(t *testing.T) {
	ctx := context.Background()
	appDeployment := validAppDeployment.DeepCopy()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorder := mockpkg.NewMockEventRecorder(mockCtrl)
	// mockStatusWriter := mockpkg.NewMockStatusWriter(mockCtrl)

	adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
	assert.NotNil(t, adapter)

	t.Run("Happy path: finalizer not present", func(t *testing.T) {
		mockClient.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		res, err := adapter.EnsureFinalizer(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{
			RequeueDelay: reconciler.DefaultRequeueDelay,
		}, res)
	})

	t.Run("Sad path: update fails", func(t *testing.T) {
		testErr := errors.New("update error")
		mockClient.EXPECT().Update(ctx, gomock.Any()).Return(testErr)
		res, err := adapter.EnsureFinalizer(ctx)
		assert.ErrorIs(t, err, testErr)
		assert.Equal(t, reconciler.OperationResult{
			RequeueDelay:   reconciler.DefaultRequeueDelay,
			RequeueRequest: false,
			CancelRequest:  false,
		}, res)
	})
}

func TestAppDeploymentAdapter_EnsureFinalizerDeleted(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	var (
		mockClient   *mockpkg.MockClient
		mockRecorder *mockpkg.MockEventRecorder
	)

	t.Run("Happy path: not triggered", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockClient = mockpkg.NewMockClient(mockCtrl)

		mockRecorder = mockpkg.NewMockEventRecorder(mockCtrl)

		appDeployment := validAppDeployment.DeepCopy()
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureFinalizerDeleted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("Happy path: finalizer deleted", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockClient = mockpkg.NewMockClient(mockCtrl)

		mockClient = mockpkg.NewMockClient(mockCtrl)
		mockRecorder = mockpkg.NewMockEventRecorder(mockCtrl)

		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Finalizers = []string{apdutil.FinalizerName}
		appDeployment.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		appDeployment.Status.Phase = apdutil.PhaseDeleted

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		mockClient.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		res, err := adapter.EnsureFinalizerDeleted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{
			RequeueDelay: reconciler.DefaultRequeueDelay,
		}, res)
	})

	t.Run("Sad path: update fails", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockClient = mockpkg.NewMockClient(mockCtrl)

		mockClient = mockpkg.NewMockClient(mockCtrl)
		mockRecorder = mockpkg.NewMockEventRecorder(mockCtrl)

		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Finalizers = []string{apdutil.FinalizerName}
		appDeployment.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		appDeployment.Status.Phase = apdutil.PhaseDeleted

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		mockClient.EXPECT().Update(ctx, gomock.Any()).Return(assert.AnError)

		res, err := adapter.EnsureFinalizerDeleted(ctx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, reconciler.OperationResult{
			RequeueDelay:   reconciler.DefaultRequeueDelay,
			RequeueRequest: false,
			CancelRequest:  false,
		}, res)
	})

	t.Run("Happy path: finalizer started but phase not in deleting", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient = mockpkg.NewMockClient(mockCtrl)
		mockEventCtrl := gomock.NewController(t)
		mockRecorder = mockpkg.NewMockEventRecorder(mockEventCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Finalizers = []string{apdutil.FinalizerName}
		appDeployment.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		appDeployment.Status.Phase = apdutil.PhasePending

		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureFinalizerDeleted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{
			RequeueDelay: reconciler.DefaultRequeueDelay,
		}, res)
	})
}

func TestAppDeploymentAdapter_EnsureDependenciesReady(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorder := mockpkg.NewMockEventRecorder(mockCtrl)

	t.Run("Happy path: skip dependencies check", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureDependenciesReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("Happy path: no dependencies ready", func(t *testing.T) {
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhasePending
		appDeployment.Spec.OpId = testOpId
		appDeployment.Spec.Dependencies = []string{
			"test-app-1",
		}
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)

		dependendApp := &v1alpha1.AppDeployment{
			Status: v1alpha1.AppDeploymentStatus{
				Phase: apdutil.PhaseReady,
			},
		}

		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf(&v1alpha1.AppDeployment{}), gomock.Any()).DoAndReturn(
			func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
				*obj.(*v1alpha1.AppDeployment) = *dependendApp
				assert.Equal(t, "test-op-id-test-app-1", key.Name)
				return nil
			}).Times(1)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		res, err := adapter.EnsureDependenciesReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: false}, res)
	})

	t.Run("Sad path: dependency not found", func(t *testing.T) {
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhasePending
		appDeployment.Spec.OpId = testOpId
		appDeployment.Spec.Dependencies = []string{
			"test-app-1",
		}
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf(&v1alpha1.AppDeployment{}), gomock.Any()).Return(assert.AnError).Times(1)

		res, err := adapter.EnsureDependenciesReady(ctx)
		assert.ErrorContains(t, err, "dependency not found: test-op-id-test-app-1")
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})

	t.Run("Sad path: dependency not ready", func(t *testing.T) {
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhasePending
		appDeployment.Spec.OpId = testOpId
		appDeployment.Spec.Dependencies = []string{
			"test-app-1",
		}
		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)

		dependendApp := &v1alpha1.AppDeployment{
			Status: v1alpha1.AppDeploymentStatus{
				Phase: apdutil.PhasePending,
			},
		}

		mockClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf(&v1alpha1.AppDeployment{}), gomock.Any()).DoAndReturn(
			func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
				*obj.(*v1alpha1.AppDeployment) = *dependendApp
				assert.Equal(t, "test-op-id-test-app-1", key.Name)
				return nil
			}).Times(1)

		res, err := adapter.EnsureDependenciesReady(ctx)
		assert.ErrorContains(t, err, "dependency is not ready: test-op-id-test-app-1")
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}

func TestAppDeploymentAdapter_EnsureDeployingFinished(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	t.Run("Happy path: skip when not in deploying phase", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)

		res, err := adapter.EnsureDeployingFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("Happy path: deploying finished", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeploying

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&batchv1.Job{})).
			DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
				*obj.(*batchv1.Job) = batchv1.Job{
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:   batchv1.JobComplete,
								Status: "True",
							},
						},
						Succeeded: 1,
					},

					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-job",
						Namespace: "default",
					},
				}
				return nil
			})
		mockClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		res, err := adapter.EnsureDeployingFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})

	t.Run("Happy path: deploying create new job", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		scheme := runtime.NewScheme()
		_ = v1alpha1.AddToScheme(scheme)
		mockClient.EXPECT().Scheme().Return(scheme).AnyTimes()

		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeploying

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&batchv1.Job{})).
			Return(k8serr.NewNotFound(batchv1.Resource("job"), "test-job"))

		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		res, err := adapter.EnsureDeployingFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})

	t.Run("Happy path: deploying job failed, create new job", func(t *testing.T) {
		failedJob := batchv1.Job{
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:   batchv1.JobFailed,
						Status: "True",
					},
				},
				Failed: 1,
			},

			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job",
				Namespace: "default",
			},
		}
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		scheme := runtime.NewScheme()
		_ = v1alpha1.AddToScheme(scheme)
		mockClient.EXPECT().Scheme().Return(scheme).AnyTimes()

		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeploying

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&batchv1.Job{})).DoAndReturn(
			func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
				*obj.(*batchv1.Job) = failedJob
				return nil
			})
		mockClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		res, err := adapter.EnsureDeployingFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}

func TestAppDeploymentAdapter_EnsureTeardownFinished(t *testing.T) {
	ctx := context.Background()

	succeededJob := batchv1.Job{
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobComplete,
					Status: "True",
				},
			},
			Succeeded: 1,
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
	}

	failedJob := batchv1.Job{
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:   batchv1.JobFailed,
					Status: "True",
				},
			},
			Failed: 1,
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
	}

	t.Run("Happy path: skip when not in deploying phase", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		logger := log.FromContext(ctx)

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorder := mockpkg.NewMockEventRecorder(mockCtrl)

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)
		res, err := adapter.EnsureTeardownFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("Happy path: teardown finished", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeleting
		logger := log.FromContext(ctx)

		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&batchv1.Job{})).
			DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
				*obj.(*batchv1.Job) = succeededJob
				return nil
			})
		mockClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockRecorder.EXPECT().Event(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		res, err := adapter.EnsureTeardownFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})
	t.Run("Happy path: teardown create new job", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeleting
		logger := log.FromContext(ctx)

		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		scheme := runtime.NewScheme()
		_ = v1alpha1.AddToScheme(scheme)
		mockClient.EXPECT().Scheme().Return(scheme).AnyTimes()

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&batchv1.Job{})).
			Return(k8serr.NewNotFound(batchv1.Resource("job"), "test-job"))

		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRecorder.EXPECT().Event(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		res, err := adapter.EnsureTeardownFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
	t.Run("Happy path: teardown job failed, create new job", func(t *testing.T) {
		appDeployment := validAppDeployment.DeepCopy()
		appDeployment.Status.Phase = apdutil.PhaseDeleting
		logger := log.FromContext(ctx)

		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()
		scheme := runtime.NewScheme()
		_ = v1alpha1.AddToScheme(scheme)
		mockClient.EXPECT().Scheme().Return(scheme).AnyTimes()

		adapter := NewAppDeploymentAdapter(ctx, appDeployment, logger, mockClient, mockRecorder)
		assert.NotNil(t, adapter)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&batchv1.Job{})).
			DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
				*obj.(*batchv1.Job) = failedJob
				return nil
			})
		mockClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockRecorder.EXPECT().Event(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		res, err := adapter.EnsureTeardownFinished(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}
