// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Azure/operation-cache-controller/internal/handler (interfaces: AppDeploymentHandlerInterface)
//
// Generated by this command:
//
//	mockgen -destination=./mocks/mock_appdeployment.go -package=mocks github.com/Azure/operation-cache-controller/internal/handler AppDeploymentHandlerInterface
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	reconciler "github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	gomock "go.uber.org/mock/gomock"
)

// MockAppDeploymentHandlerInterface is a mock of AppDeploymentHandlerInterface interface.
type MockAppDeploymentHandlerInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAppDeploymentHandlerInterfaceMockRecorder
	isgomock struct{}
}

// MockAppDeploymentHandlerInterfaceMockRecorder is the mock recorder for MockAppDeploymentHandlerInterface.
type MockAppDeploymentHandlerInterfaceMockRecorder struct {
	mock *MockAppDeploymentHandlerInterface
}

// NewMockAppDeploymentHandlerInterface creates a new mock instance.
func NewMockAppDeploymentHandlerInterface(ctrl *gomock.Controller) *MockAppDeploymentHandlerInterface {
	mock := &MockAppDeploymentHandlerInterface{ctrl: ctrl}
	mock.recorder = &MockAppDeploymentHandlerInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppDeploymentHandlerInterface) EXPECT() *MockAppDeploymentHandlerInterfaceMockRecorder {
	return m.recorder
}

// EnsureApplicationValid mocks base method.
func (m *MockAppDeploymentHandlerInterface) EnsureApplicationValid(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureApplicationValid", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureApplicationValid indicates an expected call of EnsureApplicationValid.
func (mr *MockAppDeploymentHandlerInterfaceMockRecorder) EnsureApplicationValid(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureApplicationValid", reflect.TypeOf((*MockAppDeploymentHandlerInterface)(nil).EnsureApplicationValid), ctx)
}

// EnsureDependenciesReady mocks base method.
func (m *MockAppDeploymentHandlerInterface) EnsureDependenciesReady(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureDependenciesReady", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureDependenciesReady indicates an expected call of EnsureDependenciesReady.
func (mr *MockAppDeploymentHandlerInterfaceMockRecorder) EnsureDependenciesReady(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureDependenciesReady", reflect.TypeOf((*MockAppDeploymentHandlerInterface)(nil).EnsureDependenciesReady), ctx)
}

// EnsureDeployingFinished mocks base method.
func (m *MockAppDeploymentHandlerInterface) EnsureDeployingFinished(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureDeployingFinished", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureDeployingFinished indicates an expected call of EnsureDeployingFinished.
func (mr *MockAppDeploymentHandlerInterfaceMockRecorder) EnsureDeployingFinished(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureDeployingFinished", reflect.TypeOf((*MockAppDeploymentHandlerInterface)(nil).EnsureDeployingFinished), ctx)
}

// EnsureFinalizer mocks base method.
func (m *MockAppDeploymentHandlerInterface) EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureFinalizer", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureFinalizer indicates an expected call of EnsureFinalizer.
func (mr *MockAppDeploymentHandlerInterfaceMockRecorder) EnsureFinalizer(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureFinalizer", reflect.TypeOf((*MockAppDeploymentHandlerInterface)(nil).EnsureFinalizer), ctx)
}

// EnsureFinalizerDeleted mocks base method.
func (m *MockAppDeploymentHandlerInterface) EnsureFinalizerDeleted(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureFinalizerDeleted", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureFinalizerDeleted indicates an expected call of EnsureFinalizerDeleted.
func (mr *MockAppDeploymentHandlerInterfaceMockRecorder) EnsureFinalizerDeleted(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureFinalizerDeleted", reflect.TypeOf((*MockAppDeploymentHandlerInterface)(nil).EnsureFinalizerDeleted), ctx)
}

// EnsureTeardownFinished mocks base method.
func (m *MockAppDeploymentHandlerInterface) EnsureTeardownFinished(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureTeardownFinished", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureTeardownFinished indicates an expected call of EnsureTeardownFinished.
func (mr *MockAppDeploymentHandlerInterfaceMockRecorder) EnsureTeardownFinished(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureTeardownFinished", reflect.TypeOf((*MockAppDeploymentHandlerInterface)(nil).EnsureTeardownFinished), ctx)
}
