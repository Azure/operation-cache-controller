// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Azure/operation-cache-controller/internal/controller (interfaces: RequirementAdapterInterface)
//
// Generated by this command:
//
//	mockgen -destination=./mocks/mock_requirement_adapter.go -package=mocks github.com/Azure/operation-cache-controller/internal/controller RequirementAdapterInterface
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	reconciler "github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	gomock "go.uber.org/mock/gomock"
)

// MockRequirementAdapterInterface is a mock of RequirementAdapterInterface interface.
type MockRequirementAdapterInterface struct {
	ctrl     *gomock.Controller
	recorder *MockRequirementAdapterInterfaceMockRecorder
	isgomock struct{}
}

// MockRequirementAdapterInterfaceMockRecorder is the mock recorder for MockRequirementAdapterInterface.
type MockRequirementAdapterInterfaceMockRecorder struct {
	mock *MockRequirementAdapterInterface
}

// NewMockRequirementAdapterInterface creates a new mock instance.
func NewMockRequirementAdapterInterface(ctrl *gomock.Controller) *MockRequirementAdapterInterface {
	mock := &MockRequirementAdapterInterface{ctrl: ctrl}
	mock.recorder = &MockRequirementAdapterInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRequirementAdapterInterface) EXPECT() *MockRequirementAdapterInterfaceMockRecorder {
	return m.recorder
}

// EnsureCacheExisted mocks base method.
func (m *MockRequirementAdapterInterface) EnsureCacheExisted(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureCacheExisted", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureCacheExisted indicates an expected call of EnsureCacheExisted.
func (mr *MockRequirementAdapterInterfaceMockRecorder) EnsureCacheExisted(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureCacheExisted", reflect.TypeOf((*MockRequirementAdapterInterface)(nil).EnsureCacheExisted), ctx)
}

// EnsureCachedOperationAcquired mocks base method.
func (m *MockRequirementAdapterInterface) EnsureCachedOperationAcquired(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureCachedOperationAcquired", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureCachedOperationAcquired indicates an expected call of EnsureCachedOperationAcquired.
func (mr *MockRequirementAdapterInterfaceMockRecorder) EnsureCachedOperationAcquired(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureCachedOperationAcquired", reflect.TypeOf((*MockRequirementAdapterInterface)(nil).EnsureCachedOperationAcquired), ctx)
}

// EnsureInitialized mocks base method.
func (m *MockRequirementAdapterInterface) EnsureInitialized(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureInitialized", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureInitialized indicates an expected call of EnsureInitialized.
func (mr *MockRequirementAdapterInterfaceMockRecorder) EnsureInitialized(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureInitialized", reflect.TypeOf((*MockRequirementAdapterInterface)(nil).EnsureInitialized), ctx)
}

// EnsureNotExpired mocks base method.
func (m *MockRequirementAdapterInterface) EnsureNotExpired(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureNotExpired", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureNotExpired indicates an expected call of EnsureNotExpired.
func (mr *MockRequirementAdapterInterfaceMockRecorder) EnsureNotExpired(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureNotExpired", reflect.TypeOf((*MockRequirementAdapterInterface)(nil).EnsureNotExpired), ctx)
}

// EnsureOperationReady mocks base method.
func (m *MockRequirementAdapterInterface) EnsureOperationReady(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureOperationReady", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureOperationReady indicates an expected call of EnsureOperationReady.
func (mr *MockRequirementAdapterInterfaceMockRecorder) EnsureOperationReady(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureOperationReady", reflect.TypeOf((*MockRequirementAdapterInterface)(nil).EnsureOperationReady), ctx)
}
