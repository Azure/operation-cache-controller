package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
	mockpkg "github.com/Azure/operation-cache-controller/internal/mocks"
	ctlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
	oputils "github.com/Azure/operation-cache-controller/internal/utils/controller/operation"
	rqutils "github.com/Azure/operation-cache-controller/internal/utils/controller/requirement"
	"github.com/Azure/operation-cache-controller/internal/utils/reconciler"
)

var (
	testOperationName = "test-operation"
	emptyRequirement  = &appsv1.Requirement{}
	validRequirement  = &appsv1.Requirement{
		Spec: appsv1.RequirementSpec{
			ExpireAt: time.Now().Add(time.Hour).Format(time.RFC3339),
			Template: appsv1.OperationSpec{
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
		},
	}
	validCache = &appsv1.Cache{
		Status: appsv1.CacheStatus{
			AvailableCaches: []string{"test-cache1", "test-cache2"},
		},
	}
)

func TestNewRequirementAdapter(t *testing.T) {
	t.Run("When creating a new Requirement Adapter", func(t *testing.T) {
		ctx := context.Background()
		logger := log.FromContext(ctx)

		mockCtrl := gomock.NewController(t)
		mockClient := mockpkg.NewMockClient(mockCtrl)
		mockRecorderCtrl := gomock.NewController(t)
		mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

		requirement := emptyRequirement.DeepCopy()
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		require.NotNil(t, adapter)
	})
}
func TestRequirementAdapter_EnsureNotExpired(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)

	t.Run("happy path: continue processing when expire is not set", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Spec.ExpireAt = time.Now().Add(time.Hour).Format(time.RFC3339)
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
	t.Run("happy path: continue processing when requirement is not expired", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Spec.ExpireAt = time.Now().Add(time.Hour).Format(time.RFC3339)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)

	})
	t.Run("Sad path: failed to parse expire time", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Spec.ExpireAt = "invalid-time"
		mockRecorder.EXPECT().Event(requirement, "Warning", "InvalidExpireTime", "Failed to parse expire time")

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: delete operation when expire time is in the past", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Spec.ExpireAt = time.Now().Add(-time.Hour).Format(time.RFC3339)

		mockClient.EXPECT().Delete(ctx, requirement, gomock.Any()).Return(nil)
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
	t.Run("sad path: delete operation failed", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Spec.ExpireAt = time.Now().Add(-time.Hour).Format(time.RFC3339)

		mockClient.EXPECT().Delete(ctx, requirement, gomock.Any()).Return(assert.AnError)
		mockRecorder.EXPECT().Event(requirement, "Warning", "DeleteFailed", "Failed to delete expired requirement")

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureNotExpired(ctx)
		assert.Error(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}

func TestRequirementAdapter_EnsureInitialized(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl := gomock.NewController(t)
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

	t.Run("happy path: continue processing when requirement is in empty phase and cache disabled", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureInitialized(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
	})

	t.Run("happy path: continue processing when requirement is in empty phase and cache enabled", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Spec.EnableCache = true
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureInitialized(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, rqutils.PhaseCacheChecking, requirement.Status.Phase)
	})

	t.Run("happy path: continue processing requirement is not in empty phase", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseOperating
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureInitialized(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})
}

func TestRequirementAdapter_EnsureCacheExisted(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl := gomock.NewController(t)
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

	t.Run("happy path: continue processing when cache is not enabled", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureCacheExisted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: continue processing when candidate operation exist", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		requirement.Status.OperationName = testOperationName
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureCacheExisted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
		assert.Equal(t, rqutils.PhaseCacheChecking, requirement.Status.Phase)
	})

	t.Run("happy path: when get a candidate operation", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		requirement.Status.CacheKey = ctlutils.NewCacheKeyFromApplications(requirement.Spec.Template.Applications)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		cache := validCache.DeepCopy()

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(cache), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Cache) = *cache
			return nil
		})

		mockClient.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&appsv1.Cache{})).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureCacheExisted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, rqutils.PhaseCacheChecking, requirement.Status.Phase)
	})
	t.Run("sad path: cache key is not set", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		requirement.Status.CacheKey = ""

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureCacheExisted(ctx)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
		assert.ErrorContains(t, err, "empty cache key")
	})

	t.Run("sad path: failed to get cache", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		requirement.Status.CacheKey = ctlutils.NewCacheKeyFromApplications(requirement.Spec.Template.Applications)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		cache := validCache.DeepCopy()

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(cache), gomock.Any()).Return(assert.AnError)

		res, err := adapter.EnsureCacheExisted(ctx)
		assert.Error(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})

	t.Run("happy path: cache not found, create a new cache", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		requirement.Status.CacheKey = ctlutils.NewCacheKeyFromApplications(requirement.Spec.Template.Applications)
		errCacheNotFound := apierrors.NewNotFound(schema.GroupResource{Group: "appsv1", Resource: "Cache"}, "cache not found")
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(errCacheNotFound)
		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureCacheExisted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
	})

	t.Run("happy path: cache is not available", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		requirement.Status.CacheKey = ctlutils.NewCacheKeyFromApplications(requirement.Spec.Template.Applications)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		cache := validCache.DeepCopy()
		cache.Status.AvailableCaches = nil

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(cache), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Cache) = *cache
			return nil
		})
		mockClient.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&appsv1.Cache{})).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureCacheExisted(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, requirement.Status.OperationName, "")
	})
}

func TestRequirementAdapter_EnsureCachedOperationAcquired(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)
	testRequirementUID := types.UID("test-uid")

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl := gomock.NewController(t)
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

	t.Run("happy path: continue processing when not in cache checking phase", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: continue processing when operation is not set", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = ""
		requirement.Status.Phase = rqutils.PhaseCacheChecking

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
	})

	t.Run("happy path: continue processing when operation is already acquired", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.UID = testRequirementUID
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		operation := validOperation.DeepCopy()
		operation.Annotations = map[string]string{
			oputils.AcquiredAnnotationKey: "2021-09-01T00:00:00Z",
		}
		operation.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "appsv1",
				Kind:       "Requirement",
				Name:       requirement.Name,
				UID:        requirement.UID,
			},
		}

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operation
			return nil
		})
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, rqutils.PhaseReady, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, CancelRequest: true}, res)
	})

	t.Run("happy path: continue processing when operation is acquired but other requirement", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.UID = testRequirementUID
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		operation := validOperation.DeepCopy()
		operation.Annotations = map[string]string{
			oputils.AcquiredAnnotationKey: "2021-09-01T00:00:00Z",
		}
		operation.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: "appsv1",
				Kind:       "Requirement",
				Name:       "other-requirement",
				UID:        "other-uid",
			},
		}

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operation
			return nil
		})
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})

	t.Run("happy path: continue processing when operation is not acquired, acquired it with success", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.UID = testRequirementUID
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		operation := validOperation.DeepCopy()
		operation.Annotations = map[string]string{}

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operation
			return nil
		})
		mockClient.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&appsv1.Operation{})).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
		assert.Equal(t, operation.Name, requirement.Status.OperationName)
		assert.Equal(t, rqutils.PhaseReady, requirement.Status.Phase)
	})

	t.Run("sad path: failed to get operation", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.UID = testRequirementUID
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).Return(assert.AnError)

		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})

	t.Run("sad path: when operation is not acquired, acquired it with failed", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.UID = testRequirementUID
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseCacheChecking
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		operation := validOperation.DeepCopy()
		operation.Annotations = map[string]string{}

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operation
			return nil
		})
		mockClient.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&appsv1.Operation{})).Return(assert.AnError)

		res, err := adapter.EnsureCachedOperationAcquired(ctx)
		assert.Error(t, err)
		assert.Equal(t, requirement.Status.OperationName, operation.Name)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})
}

func TestRequirementAdapter_EnsureOperationReady(t *testing.T) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	mockCtrl := gomock.NewController(t)
	mockClient := mockpkg.NewMockClient(mockCtrl)
	mockRecorderCtrl := gomock.NewController(t)
	mockRecorder := mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl := gomock.NewController(t)
	mockStatusWriter := mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	mockClient.EXPECT().Status().Return(mockStatusWriter).AnyTimes()

	t.Run("happy path: continue processing when not in ready and operating phase", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: continue processing when in ready phase but cachekey is not changed", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.CacheKey = ctlutils.NewCacheKeyFromApplications(requirement.Spec.Template.Applications)
		requirement.Status.Phase = rqutils.PhaseReady
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, reconciler.OperationResult{}, res)
	})

	t.Run("happy path: ready phase but cachekey is changed", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseReady
		requirement.Status.CacheKey = "test-cache-key"
		operaition := validOperation.DeepCopy()

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operaition
			return nil
		})
		mockClient.EXPECT().Update(ctx, gomock.AssignableToTypeOf(&appsv1.Operation{})).Return(nil)
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		res, err := adapter.EnsureOperationReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})
	t.Run("sad path: failed to get operation", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseReady
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).Return(assert.AnError)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, rqutils.PhaseReady, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})

	t.Run("happy path: continue processing when operation is not ready", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseOperating

		operation := validOperation.DeepCopy()
		operation.Status.Phase = oputils.PhaseReconciling

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operation
			return nil
		})

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})

	t.Run("happy path: continue processing when operation is ready", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseOperating
		operation := validOperation.DeepCopy()
		operation.Status.Phase = oputils.PhaseReconciled

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).DoAndReturn(func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			*obj.(*appsv1.Operation) = *operation
			return nil
		})
		mockStatusWriter.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, rqutils.PhaseReady, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay}, res)
	})

	t.Run("happy path: operation not found, create one", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseOperating
		scheme := runtime.NewScheme()
		_ = appsv1.AddToScheme(scheme)
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)

		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).Return(apierrors.NewNotFound(schema.GroupResource{Group: "appsv1", Resource: "Operation"}, "operation not found"))
		mockClient.EXPECT().Scheme().Return(scheme)
		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.NoError(t, err)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})

	t.Run("sad path: failed to create operation", func(t *testing.T) {
		requirement := validRequirement.DeepCopy()
		requirement.Status.OperationName = testOperationName
		requirement.Status.Phase = rqutils.PhaseOperating
		adapter := NewRequirementAdapter(ctx, requirement, logger, mockClient, mockRecorder)
		schema := runtime.NewScheme()
		_ = appsv1.AddToScheme(schema)
		mockClient.EXPECT().Get(ctx, gomock.Any(), gomock.AssignableToTypeOf(&appsv1.Operation{}), gomock.Any()).Return(assert.AnError)
		mockClient.EXPECT().Scheme().Return(schema)
		mockClient.EXPECT().Create(ctx, gomock.Any()).Return(assert.AnError)

		res, err := adapter.EnsureOperationReady(ctx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Equal(t, rqutils.PhaseOperating, requirement.Status.Phase)
		assert.Equal(t, reconciler.OperationResult{RequeueDelay: reconciler.DefaultRequeueDelay, RequeueRequest: true}, res)
	})
}
