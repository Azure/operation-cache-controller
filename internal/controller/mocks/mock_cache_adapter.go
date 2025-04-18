// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Azure/operation-cache-controller/internal/controller (interfaces: CacheAdapterInterface)
//
// Generated by this command:
//
//	mockgen -destination=./mocks/mock_cache_adapter.go -package=mocks github.com/Azure/operation-cache-controller/internal/controller CacheAdapterInterface
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	reconciler "github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	gomock "go.uber.org/mock/gomock"
)

// MockCacheAdapterInterface is a mock of CacheAdapterInterface interface.
type MockCacheAdapterInterface struct {
	ctrl     *gomock.Controller
	recorder *MockCacheAdapterInterfaceMockRecorder
	isgomock struct{}
}

// MockCacheAdapterInterfaceMockRecorder is the mock recorder for MockCacheAdapterInterface.
type MockCacheAdapterInterfaceMockRecorder struct {
	mock *MockCacheAdapterInterface
}

// NewMockCacheAdapterInterface creates a new mock instance.
func NewMockCacheAdapterInterface(ctrl *gomock.Controller) *MockCacheAdapterInterface {
	mock := &MockCacheAdapterInterface{ctrl: ctrl}
	mock.recorder = &MockCacheAdapterInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheAdapterInterface) EXPECT() *MockCacheAdapterInterfaceMockRecorder {
	return m.recorder
}

// AdjustCache mocks base method.
func (m *MockCacheAdapterInterface) AdjustCache(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AdjustCache", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AdjustCache indicates an expected call of AdjustCache.
func (mr *MockCacheAdapterInterfaceMockRecorder) AdjustCache(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AdjustCache", reflect.TypeOf((*MockCacheAdapterInterface)(nil).AdjustCache), ctx)
}

// CalculateKeepAliveCount mocks base method.
func (m *MockCacheAdapterInterface) CalculateKeepAliveCount(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculateKeepAliveCount", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CalculateKeepAliveCount indicates an expected call of CalculateKeepAliveCount.
func (mr *MockCacheAdapterInterfaceMockRecorder) CalculateKeepAliveCount(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculateKeepAliveCount", reflect.TypeOf((*MockCacheAdapterInterface)(nil).CalculateKeepAliveCount), ctx)
}

// CheckCacheExpiry mocks base method.
func (m *MockCacheAdapterInterface) CheckCacheExpiry(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckCacheExpiry", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckCacheExpiry indicates an expected call of CheckCacheExpiry.
func (mr *MockCacheAdapterInterfaceMockRecorder) CheckCacheExpiry(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckCacheExpiry", reflect.TypeOf((*MockCacheAdapterInterface)(nil).CheckCacheExpiry), ctx)
}

// EnsureCacheInitialized mocks base method.
func (m *MockCacheAdapterInterface) EnsureCacheInitialized(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureCacheInitialized", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureCacheInitialized indicates an expected call of EnsureCacheInitialized.
func (mr *MockCacheAdapterInterfaceMockRecorder) EnsureCacheInitialized(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureCacheInitialized", reflect.TypeOf((*MockCacheAdapterInterface)(nil).EnsureCacheInitialized), ctx)
}
