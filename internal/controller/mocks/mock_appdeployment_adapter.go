// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Azure/operation-cache-controller/internal/controller (interfaces: AppDeploymentAdapterInterface)
//
// Generated by this command:
//
//	mockgen -destination=./mocks/mock_appdeployment_adapter.go -package=mocks github.com/Azure/operation-cache-controller/internal/controller AppDeploymentAdapterInterface
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	reconciler "github.com/Azure/operation-cache-controller/internal/utils/reconciler"
	gomock "go.uber.org/mock/gomock"
)

// MockAppDeploymentAdapterInterface is a mock of AppDeploymentAdapterInterface interface.
type MockAppDeploymentAdapterInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAppDeploymentAdapterInterfaceMockRecorder
	isgomock struct{}
}

// MockAppDeploymentAdapterInterfaceMockRecorder is the mock recorder for MockAppDeploymentAdapterInterface.
type MockAppDeploymentAdapterInterfaceMockRecorder struct {
	mock *MockAppDeploymentAdapterInterface
}

// NewMockAppDeploymentAdapterInterface creates a new mock instance.
func NewMockAppDeploymentAdapterInterface(ctrl *gomock.Controller) *MockAppDeploymentAdapterInterface {
	mock := &MockAppDeploymentAdapterInterface{ctrl: ctrl}
	mock.recorder = &MockAppDeploymentAdapterInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppDeploymentAdapterInterface) EXPECT() *MockAppDeploymentAdapterInterfaceMockRecorder {
	return m.recorder
}

// EnsureApplicationValid mocks base method.
func (m *MockAppDeploymentAdapterInterface) EnsureApplicationValid(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureApplicationValid", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureApplicationValid indicates an expected call of EnsureApplicationValid.
func (mr *MockAppDeploymentAdapterInterfaceMockRecorder) EnsureApplicationValid(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureApplicationValid", reflect.TypeOf((*MockAppDeploymentAdapterInterface)(nil).EnsureApplicationValid), ctx)
}

// EnsureDependenciesReady mocks base method.
func (m *MockAppDeploymentAdapterInterface) EnsureDependenciesReady(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureDependenciesReady", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureDependenciesReady indicates an expected call of EnsureDependenciesReady.
func (mr *MockAppDeploymentAdapterInterfaceMockRecorder) EnsureDependenciesReady(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureDependenciesReady", reflect.TypeOf((*MockAppDeploymentAdapterInterface)(nil).EnsureDependenciesReady), ctx)
}

// EnsureDeployingFinished mocks base method.
func (m *MockAppDeploymentAdapterInterface) EnsureDeployingFinished(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureDeployingFinished", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureDeployingFinished indicates an expected call of EnsureDeployingFinished.
func (mr *MockAppDeploymentAdapterInterfaceMockRecorder) EnsureDeployingFinished(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureDeployingFinished", reflect.TypeOf((*MockAppDeploymentAdapterInterface)(nil).EnsureDeployingFinished), ctx)
}

// EnsureFinalizer mocks base method.
func (m *MockAppDeploymentAdapterInterface) EnsureFinalizer(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureFinalizer", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureFinalizer indicates an expected call of EnsureFinalizer.
func (mr *MockAppDeploymentAdapterInterfaceMockRecorder) EnsureFinalizer(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureFinalizer", reflect.TypeOf((*MockAppDeploymentAdapterInterface)(nil).EnsureFinalizer), ctx)
}

// EnsureFinalizerDeleted mocks base method.
func (m *MockAppDeploymentAdapterInterface) EnsureFinalizerDeleted(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureFinalizerDeleted", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureFinalizerDeleted indicates an expected call of EnsureFinalizerDeleted.
func (mr *MockAppDeploymentAdapterInterfaceMockRecorder) EnsureFinalizerDeleted(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureFinalizerDeleted", reflect.TypeOf((*MockAppDeploymentAdapterInterface)(nil).EnsureFinalizerDeleted), ctx)
}

// EnsureTeardownFinished mocks base method.
func (m *MockAppDeploymentAdapterInterface) EnsureTeardownFinished(ctx context.Context) (reconciler.OperationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureTeardownFinished", ctx)
	ret0, _ := ret[0].(reconciler.OperationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EnsureTeardownFinished indicates an expected call of EnsureTeardownFinished.
func (mr *MockAppDeploymentAdapterInterfaceMockRecorder) EnsureTeardownFinished(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureTeardownFinished", reflect.TypeOf((*MockAppDeploymentAdapterInterface)(nil).EnsureTeardownFinished), ctx)
}
