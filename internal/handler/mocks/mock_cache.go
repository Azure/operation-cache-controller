// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Azure/operation-cache-controller/internal/handler (interfaces: CacheHandlerInterface)
//
// Generated by this command:
//
//	mockgen -destination=./mocks/mock_cache.go -package=mocks github.com/Azure/operation-cache-controller/internal/handler CacheHandlerInterface
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	reconciler "github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	gomock "go.uber.org/mock/gomock"
)

// MockCacheHandlerInterface is a mock of CacheHandlerInterface interface.
type MockCacheHandlerInterface struct {
	ctrl     *gomock.Controller
	recorder *MockCacheHandlerInterfaceMockRecorder
	isgomock struct{}
}

// MockCacheHandlerInterfaceMockRecorder is the mock recorder for MockCacheHandlerInterface.
type MockCacheHandlerInterfaceMockRecorder struct {
	mock *MockCacheHandlerInterface
}

// NewMockCacheHandlerInterface creates a new mock instance.
func NewMockCacheHandlerInterface(ctrl *gomock.Controller) *MockCacheHandlerInterface {
	mock := &MockCacheHandlerInterface{ctrl: ctrl}
	mock.recorder = &MockCacheHandlerInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacheHandlerInterface) EXPECT() *MockCacheHandlerInterfaceMockRecorder {
	return m.recorder
}

// AdjustCache mocks base method.
func (m *MockCacheHandlerInterface) AdjustCache(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AdjustCache", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AdjustCache indicates an expected call of AdjustCache.
func (mr *MockCacheHandlerInterfaceMockRecorder) AdjustCache(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AdjustCache", reflect.TypeOf((*MockCacheHandlerInterface)(nil).AdjustCache), ctx)
}

// CalculateKeepAliveCount mocks base method.
func (m *MockCacheHandlerInterface) CalculateKeepAliveCount(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CalculateKeepAliveCount", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CalculateKeepAliveCount indicates an expected call of CalculateKeepAliveCount.
func (mr *MockCacheHandlerInterfaceMockRecorder) CalculateKeepAliveCount(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CalculateKeepAliveCount", reflect.TypeOf((*MockCacheHandlerInterface)(nil).CalculateKeepAliveCount), ctx)
}

// CheckCacheExpiry mocks base method.
func (m *MockCacheHandlerInterface) CheckCacheExpiry(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckCacheExpiry", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CheckCacheExpiry indicates an expected call of CheckCacheExpiry.
func (mr *MockCacheHandlerInterfaceMockRecorder) CheckCacheExpiry(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckCacheExpiry", reflect.TypeOf((*MockCacheHandlerInterface)(nil).CheckCacheExpiry), ctx)
}

// EnsureCacheInitialized mocks base method.
func (m *MockCacheHandlerInterface) EnsureCacheInitialized(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureCacheInitialized", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureCacheInitialized indicates an expected call of EnsureCacheInitialized.
func (mr *MockCacheHandlerInterfaceMockRecorder) EnsureCacheInitialized(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureCacheInitialized", reflect.TypeOf((*MockCacheHandlerInterface)(nil).EnsureCacheInitialized), ctx)
}
