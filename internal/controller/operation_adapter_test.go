package controller

import (
	"context"
	"testing"
	"time"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
	mockpkg "github.com/Azure/operation-cache-controller/internal/mocks"
	adputils "github.com/Azure/operation-cache-controller/internal/utils/controller/appdeployment"
	oputils "github.com/Azure/operation-cache-controller/internal/utils/controller/operation"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	emptyOperation = &appsv1.Operation{}
	validOperation = &appsv1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-operation",
			Namespace: "default",
		},
		Spec: appsv1.OperationSpec{
			Applications: []appsv1.ApplicationSpec{
				{
					Name:      "test-app1",
					Provision: newTestJobSpec(),
					Teardown:  newTestJobSpec(),

					Dependencies: []string{"test-app2"},
				},
				{
					Name:      "test-app2",
					Provision: newTestJobSpec(),
					Teardown:  newTestJobSpec(),
				},
			},
		},
	}
	emptyAppDeploymentList = &appsv1.AppDeploymentList{}

	validAppDeploymentList = &appsv1.AppDeploymentList{
		Items: []appsv1.AppDeployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app1",
					Namespace: "default",
				},
				Spec: appsv1.AppDeploymentSpec{
					OpId:         "test-operation",
					Provision:    newTestJobSpec(),
					Teardown:     newTestJobSpec(),
					Dependencies: []string{"test-app2"},
				},
				Status: appsv1.AppDeploymentStatus{
					Phase: adputils.PhaseReady,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app2",
					Namespace: "default",
				},
				Spec: appsv1.AppDeploymentSpec{
					OpId:      "test-operation",
					Provision: newTestJobSpec(),
					Teardown:  newTestJobSpec(),
				},
				Status: appsv1.AppDeploymentStatus{
					Phase: adputils.PhaseReady,
				},
			},
		},
	}

	changedValidAppDeploymentList = &appsv1.AppDeploymentList{
		Items: []appsv1.AppDeployment{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app2",
					Namespace: "default",
				},
				Spec: appsv1.AppDeploymentSpec{
					OpId: "test-operation",
					Provision: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "test-container",
										Image: "test-image",
										Command: []string{
											"echo",
											"world",
										},
										Args: []string{
											"hello",
										},
									},
								},
							},
						},
					},
					Teardown: newTestJobSpec(),
				},
				Status: appsv1.AppDeploymentStatus{
					Phase: adputils.PhaseReady,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-app3",
					Namespace: "default",
				},
				Spec: appsv1.AppDeploymentSpec{
					OpId:      "test-operation",
					Provision: newTestJobSpec(),
					Teardown:  newTestJobSpec(),
				},
				Status: appsv1.AppDeploymentStatus{
					Phase: adputils.PhaseReady,
				},
			},
		},
	}
)

func TestNewOperationAdapter(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

	operation := emptyOperation.DeepCopy()
	adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
	assert.NotNil(t, adapter)
}

func TestOperationAdapter_EnsureNotExpired(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

	t.Run("happy path: continue processing when expire is not set", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Spec.ExpireAt = time.Now().Add(time.Hour).Format(time.RFC3339)
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
	t.Run("happy path: continue processing when operation is in deleting phase", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Spec.ExpireAt = time.Now().Add(time.Hour).Format(time.RFC3339)
		operation.Status.Phase = oputils.PhaseDeleting

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)

		operation.Status.Phase = oputils.PhaseDeleted
		adapter = NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err = adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
	t.Run("happy path: continue processing when expire time is in the future", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Spec.ExpireAt = time.Now().Add(time.Hour).Format(time.RFC3339)
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
	t.Run("Sad path: failed to parse expire time", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Spec.ExpireAt = "invalid-time"
		mockRecorder.EXPECT().Event(operation, "Warning", "InvalidExpireTime", "Failed to parse expire time")

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: delete operation when expire time is in the past", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Spec.ExpireAt = time.Now().Add(-time.Hour).Format(time.RFC3339)

		mockClient.EXPECT().Delete(ctx, operation, gomock.Any()).Return(nil)

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
	t.Run("sad path: delete operation failed", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Spec.ExpireAt = time.Now().Add(-time.Hour).Format(time.RFC3339)

		mockClient.EXPECT().Delete(ctx, operation, gomock.Any()).Return(assert.AnError)

		mockRecorder.EXPECT().Event(operation, "Warning", "DeleteFailed", "Failed to delete expired operation")

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.Error(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}

func TestOperationAdapter_EnsureAllAppsAreReady(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	t.Run("happy path: continue processing when operation is in deleting phase", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

		operation := validOperation.DeepCopy()
		operation.Status.Phase = oputils.PhaseDeleting

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureAllAppsAreReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)

		operation.Status.Phase = oputils.PhaseDeleted
		adapter = NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err = adapter.EnsureAllAppsAreReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: continue processing when operation is in empty phase", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockStatusWriterCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		operation := emptyOperation.DeepCopy()

		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureAllAppsAreReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.NotEmpty(t, operation.Status.CacheKey)
		assert.Equal(t, operation.Status.Phase, oputils.PhaseReconciling)
	})

	t.Run("happy path: continue processing when operation is in reconciling phase", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

		operation := validOperation.DeepCopy()
		operation.Status.Phase = oputils.PhaseReconciling

		appList := emptyAppDeploymentList.DeepCopy()
		mockClient.EXPECT().List(ctx, appList, gomock.Any()).DoAndReturn(func(ctx context.Context, list *appsv1.AppDeploymentList, opts ...interface{}) error {
			*list = *changedValidAppDeploymentList
			return nil
		})
		scheme := runtime.NewScheme()
		utilruntime.Must(appsv1.AddToScheme(scheme))
		mockClient.EXPECT().Scheme().Return(scheme).AnyTimes()
		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opt ...interface{}) error {
			*obj.(*appsv1.AppDeployment) = appsv1.AppDeployment{}
			return nil
		}).AnyTimes()
		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockClient.EXPECT().Delete(ctx, gomock.Any(), gomock.Any()).Return(nil)
		mockClient.EXPECT().Update(ctx, gomock.Any()).Return(nil)
		mockRecorder.EXPECT().Event(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureAllAppsAreReady(ctx)
		assert.ErrorContains(t, err, "app deployment is not ready")
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
		assert.Equal(t, operation.Status.Phase, oputils.PhaseReconciling)

	})

	t.Run("happy path: continue processing when all apps are ready", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		mockStatusWriterCtrl := gomock.NewController(t)
		mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
		mockClient.EXPECT().Status().Return(mockStatusWriter)

		// set operation to reconciling phase
		operation := validOperation.DeepCopy()
		operation.Status.Phase = oputils.PhaseReconciling

		appList := emptyAppDeploymentList.DeepCopy()
		mockClient.EXPECT().List(ctx, appList, gomock.Any()).DoAndReturn(func(ctx context.Context, list *appsv1.AppDeploymentList, opts ...interface{}) error {
			*list = *validAppDeploymentList
			return nil
		})
		scheme := runtime.NewScheme()
		mockClient.EXPECT().Scheme().Return(scheme).AnyTimes()
		readyAppDeployment := &appsv1.AppDeployment{}
		readyAppDeployment.Status.Phase = adputils.PhaseReady
		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj runtime.Object, opt ...interface{}) error {
			*obj.(*appsv1.AppDeployment) = *readyAppDeployment
			return nil
		}).AnyTimes()
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureAllAppsAreReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, CancelRequest: true}, res)
		assert.Equal(t, operation.Status.Phase, oputils.PhaseReconciled)
	})

}

func TestOperationAdapter_EnsureFinalizer(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

	t.Run("happy path: continue processing when finalizer is not set", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.ObjectMeta.Finalizers = nil
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Update(ctx, operation).Return(nil)
		res, err := adapter.EnsureFinalizer(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})
}

func TestOperationAdapter_EnsureFinalizerRemoved(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl := gomock.NewController(t)
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	mockClient.EXPECT().Status().Return(mockStatusWriter)

	t.Run("happy path: continue processing when finalizer is not set", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.ObjectMeta.Finalizers = nil
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureFinalizerRemoved(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: continue processing when operation is in deleted phase", func(t *testing.T) {

		operation := validOperation.DeepCopy()
		operation.ObjectMeta.Finalizers = []string{oputils.FinalizerName}
		operation.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		operation.Status.Phase = oputils.PhaseDeleted
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Update(ctx, operation).Return(nil)

		res, err := adapter.EnsureFinalizerRemoved(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, len(operation.ObjectMeta.Finalizers), 0)
	})

	t.Run("happy path: continue processing when operation is not in deleting phase", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.ObjectMeta.Finalizers = []string{oputils.FinalizerName}
		operation.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)

		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureFinalizerRemoved(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})
}

func TestOperationAdapter_EnsureAllAppsAreDeleted(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl := gomock.NewController(t)
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	mockClient.EXPECT().Status().Return(mockStatusWriter)

	t.Run("happy path: continue processing when operation is in deleting phase", func(t *testing.T) {
		operation := validOperation.DeepCopy()
		operation.Status.Phase = oputils.PhaseDeleting
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)

		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		res, err := adapter.EnsureAllAppsAreDeleted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, CancelRequest: true}, res)
		assert.Equal(t, operation.Status.Phase, oputils.PhaseDeleted)

	})

	t.Run("happy path: continue processing when operation is in empty phase", func(t *testing.T) {
		operation := emptyOperation.DeepCopy()
		adapter := NewOperationAdapter(ctx, operation, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureAllAppsAreDeleted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
}
