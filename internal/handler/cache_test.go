package handler

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	ctrlutils "github.com/Azure/operation-cache-controller/internal/utils/controller"
	mockpkg "github.com/Azure/operation-cache-controller/internal/utils/mocks"
)

var cacheHelper = ctrlutils.NewCacheHelper()

func TestNewCacheAdapter(t *testing.T) {
	t.Run("NewCacheAdapter", func(t *testing.T) {
		testCache := &v1alpha1.Cache{}
		testlogger := logr.Logger{}
		scheme := runtime.NewScheme()
		var (
			mockClientCtrl   *gomock.Controller
			mockClient       *mockpkg.MockClient
			mockRecorderCtrl *gomock.Controller
			mockRecorder     *mockpkg.MockEventRecorder
		)
		mockClient = mockpkg.NewMockClient(mockClientCtrl)
		mockRecorder = mockpkg.NewMockEventRecorder(mockRecorderCtrl)
		adapter := NewCacheHandler(context.Background(), testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
		assert.NotNil(t, adapter)
	})
}

func getTestApps() []v1alpha1.ApplicationSpec {
	return []v1alpha1.ApplicationSpec{
		{
			Name: "test-app-available",
			Provision: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{}},
					},
				},
			},
		},
	}
}

func TestCacheCheckCacheExpiry(t *testing.T) {
	ctx := context.Background()
	testlogger := log.FromContext(ctx)
	scheme := runtime.NewScheme()
	var (
		mockClientCtrl   *gomock.Controller
		mockClient       *mockpkg.MockClient
		mockRecorderCtrl *gomock.Controller
		mockRecorder     *mockpkg.MockEventRecorder
	)
	mockClientCtrl = gomock.NewController(t)
	mockRecorderCtrl = gomock.NewController(t)
	mockClient = mockpkg.NewMockClient(mockClientCtrl)
	mockRecorder = mockpkg.NewMockEventRecorder(mockRecorderCtrl)

	t.Run("happy path", func(t *testing.T) {
		t.Run("cache not expired", func(t *testing.T) {
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.CacheSpec{
					ExpireTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
				Status: v1alpha1.CacheStatus{},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
			assert.NotNil(t, adapter)

			res, err := adapter.CheckCacheExpiry(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, false, res.CancelRequest)
		})
		t.Run("cache expired", func(t *testing.T) {
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.CacheSpec{
					ExpireTime: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
				Status: v1alpha1.CacheStatus{},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
			assert.NotNil(t, adapter)
			mockClient.EXPECT().Delete(ctx, gomock.Any()).Return(nil)

			res, err := adapter.CheckCacheExpiry(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, true, res.CancelRequest)
		})
		t.Run("cache expireTime not set", func(t *testing.T) {
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec:   v1alpha1.CacheSpec{},
				Status: v1alpha1.CacheStatus{},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
			assert.NotNil(t, adapter)

			res, err := adapter.CheckCacheExpiry(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, false, res.CancelRequest)
		})
	})

	t.Run("negative cases", func(t *testing.T) {
		t.Run("invalid expire time", func(t *testing.T) {
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.CacheSpec{
					ExpireTime: "invalid-time",
				},
				Status: v1alpha1.CacheStatus{},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
			assert.NotNil(t, adapter)

			res, err := adapter.CheckCacheExpiry(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, false, res.CancelRequest)
		})
	})
}

func TestCacheEnsureCacheInitialized(t *testing.T) {
	ctx := context.Background()
	testlogger := log.FromContext(ctx)
	scheme := runtime.NewScheme()
	var (
		mockClientCtrl       *gomock.Controller
		mockClient           *mockpkg.MockClient
		mockRecorderCtrl     *gomock.Controller
		mockRecorder         *mockpkg.MockEventRecorder
		mockStatusWriterCtrl *gomock.Controller
		mockStatusWriter     *mockpkg.MockStatusWriter
	)
	mockClientCtrl = gomock.NewController(t)
	mockRecorderCtrl = gomock.NewController(t)
	mockClient = mockpkg.NewMockClient(mockClientCtrl)
	mockRecorder = mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl = gomock.NewController(t)
	mockStatusWriter = mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)

	testApps := getTestApps()
	testCacheKey := cacheHelper.NewCacheKeyFromApplications(testApps)

	t.Run("happy path", func(t *testing.T) {
		testCache := &v1alpha1.Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cache",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CacheSpec{
				OperationTemplate: v1alpha1.OperationSpec{
					Applications: testApps,
				},
			},
			Status: v1alpha1.CacheStatus{},
		}
		adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
		assert.NotNil(t, adapter)
		mockClient.EXPECT().Status().Return(mockStatusWriter)
		mockStatusWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		res, err := adapter.EnsureCacheInitialized(ctx)
		assert.Nil(t, err)
		assert.Equal(t, false, res.RequeueRequest)
		assert.Equal(t, testCache.Status.CacheKey, testCacheKey)
	})
}

func TestCacheCalculateKeepAliveCount(t *testing.T) {
	ctx := context.Background()
	testlogger := log.FromContext(ctx)
	scheme := runtime.NewScheme()
	var (
		mockClientCtrl       *gomock.Controller
		mockClient           *mockpkg.MockClient
		mockRecorderCtrl     *gomock.Controller
		mockRecorder         *mockpkg.MockEventRecorder
		mockStatusWriterCtrl *gomock.Controller
		mockStatusWriter     *mockpkg.MockStatusWriter
	)
	mockClientCtrl = gomock.NewController(t)
	mockRecorderCtrl = gomock.NewController(t)
	mockClient = mockpkg.NewMockClient(mockClientCtrl)
	mockRecorder = mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl = gomock.NewController(t)
	mockStatusWriter = mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)
	testApps := getTestApps()

	t.Run("happy path", func(t *testing.T) {
		testCache := &v1alpha1.Cache{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cache",
				Namespace: "test-ns",
			},
			Spec: v1alpha1.CacheSpec{
				OperationTemplate: v1alpha1.OperationSpec{
					Applications: testApps,
				},
			},
			Status: v1alpha1.CacheStatus{},
		}
		adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
		assert.NotNil(t, adapter)
		mockClient.EXPECT().Status().Return(mockStatusWriter)
		mockStatusWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

		res, err := adapter.CalculateKeepAliveCount(ctx)
		assert.Nil(t, err)
		assert.Equal(t, false, res.RequeueRequest)
		assert.Equal(t, testCache.Status.KeepAliveCount, int32(5))
	})
}

func TestCacheAdjustCache(t *testing.T) {
	ctx := context.Background()
	testlogger := log.FromContext(ctx)
	scheme := runtime.NewScheme()
	var (
		mockClientCtrl       *gomock.Controller
		mockClient           *mockpkg.MockClient
		mockRecorderCtrl     *gomock.Controller
		mockRecorder         *mockpkg.MockEventRecorder
		mockStatusWriterCtrl *gomock.Controller
		mockStatusWriter     *mockpkg.MockStatusWriter
	)
	mockClientCtrl = gomock.NewController(t)
	mockRecorderCtrl = gomock.NewController(t)
	mockClient = mockpkg.NewMockClient(mockClientCtrl)
	mockRecorder = mockpkg.NewMockEventRecorder(mockRecorderCtrl)
	mockStatusWriterCtrl = gomock.NewController(t)
	mockStatusWriter = mockpkg.NewMockStatusWriter(mockStatusWriterCtrl)

	testApps := getTestApps()
	testCacheKey := cacheHelper.NewCacheKeyFromApplications(testApps)

	newOperation := &v1alpha1.Operation{
		Spec: v1alpha1.OperationSpec{
			Applications: testApps,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-operation-new",
			Namespace: "test-ns",
			Labels:    map[string]string{ctrlutils.LabelNameCacheKey: testCacheKey},
		},
		Status: v1alpha1.OperationStatus{
			Phase: v1alpha1.OperationPhaseEmpty,
		},
	}

	availableOperation := &v1alpha1.Operation{
		Spec: v1alpha1.OperationSpec{
			Applications: testApps,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-operation-available",
			Namespace: "test-ns",
			Labels:    map[string]string{ctrlutils.LabelNameCacheKey: testCacheKey},
		},
		Status: v1alpha1.OperationStatus{
			Phase: v1alpha1.OperationPhaseReconciled,
		},
	}
	// TODO assert operation status
	// assert.True(t, oputils.IsAvailable(ctx, availableOperation))

	// TODO add operations in other statuses

	t.Run("cache balance = 0", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			resOperations := v1alpha1.OperationList{Items: []v1alpha1.Operation{
				*newOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
			}}
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.CacheSpec{
					OperationTemplate: v1alpha1.OperationSpec{
						Applications: testApps,
					},
				},
				Status: v1alpha1.CacheStatus{
					CacheKey:       testCacheKey,
					KeepAliveCount: 2,
				},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
			assert.NotNil(t, adapter)
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).SetArg(1, resOperations).Return(nil)
			mockClient.EXPECT().Status().Return(mockStatusWriter)
			mockStatusWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

			res, err := adapter.AdjustCache(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, testCache.Status.AvailableCaches, []string{"test-operation-available", "test-operation-available"})
		})
	})

	t.Run("cache balance > 0", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			resOperationItems := []v1alpha1.Operation{
				*newOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
			}
			resOperations := v1alpha1.OperationList{Items: resOperationItems}
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.CacheSpec{
					OperationTemplate: v1alpha1.OperationSpec{
						Applications: testApps,
					},
				},
				Status: v1alpha1.CacheStatus{
					CacheKey:       testCacheKey,
					KeepAliveCount: 2,
				},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, ctrl.SetControllerReference)
			assert.NotNil(t, adapter)
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).SetArg(1, resOperations).Return(nil)
			mockClient.EXPECT().Delete(ctx, gomock.Any()).Return(nil).Times(3)
			mockClient.EXPECT().Status().Return(mockStatusWriter)
			mockStatusWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

			res, err := adapter.AdjustCache(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, testCache.Status.AvailableCaches, []string{"test-operation-available", "test-operation-available", "test-operation-available", "test-operation-available"})
		})
	})

	t.Run("cache balance < 0", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			resOperations := v1alpha1.OperationList{Items: []v1alpha1.Operation{
				*newOperation.DeepCopy(),
				*availableOperation.DeepCopy(),
			}}
			testCache := &v1alpha1.Cache{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cache",
					Namespace: "test-ns",
				},
				Spec: v1alpha1.CacheSpec{
					OperationTemplate: v1alpha1.OperationSpec{
						Applications: testApps,
					},
				},
				Status: v1alpha1.CacheStatus{
					CacheKey:       testCacheKey,
					KeepAliveCount: 3,
				},
			}
			adapter := NewCacheHandler(ctx, testCache, testlogger, mockClient, scheme, mockRecorder, func(owner, controlled metav1.Object, scheme *runtime.Scheme, opts ...controllerutil.OwnerReferenceOption) error {
				return nil
			})
			assert.NotNil(t, adapter)
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).SetArg(1, resOperations).Return(nil)
			mockClient.EXPECT().Create(ctx, gomock.Any()).Return(nil).Times(1)
			mockClient.EXPECT().Status().Return(mockStatusWriter)
			mockStatusWriter.EXPECT().Update(ctx, gomock.Any()).Return(nil)

			res, err := adapter.AdjustCache(ctx)
			assert.Nil(t, err)
			assert.Equal(t, false, res.RequeueRequest)
			assert.Equal(t, testCache.Status.AvailableCaches, []string{"test-operation-available"})
		})
	})
}
