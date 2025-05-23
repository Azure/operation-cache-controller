// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/controller-runtime/pkg/client (interfaces: StatusWriter)
//
// Generated by this command:
//
//	mockgen -destination ./mock_cr_status_writer.go -package mocks sigs.k8s.io/controller-runtime/pkg/client StatusWriter
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockStatusWriter is a mock of StatusWriter interface.
type MockStatusWriter struct {
	ctrl     *gomock.Controller
	recorder *MockStatusWriterMockRecorder
	isgomock struct{}
}

// MockStatusWriterMockRecorder is the mock recorder for MockStatusWriter.
type MockStatusWriterMockRecorder struct {
	mock *MockStatusWriter
}

// NewMockStatusWriter creates a new mock instance.
func NewMockStatusWriter(ctrl *gomock.Controller) *MockStatusWriter {
	mock := &MockStatusWriter{ctrl: ctrl}
	mock.recorder = &MockStatusWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStatusWriter) EXPECT() *MockStatusWriterMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockStatusWriter) Create(ctx context.Context, obj, subResource client.Object, opts ...client.SubResourceCreateOption) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, obj, subResource}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Create", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockStatusWriterMockRecorder) Create(ctx, obj, subResource any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, obj, subResource}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockStatusWriter)(nil).Create), varargs...)
}

// Patch mocks base method.
func (m *MockStatusWriter) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, obj, patch}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Patch", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Patch indicates an expected call of Patch.
func (mr *MockStatusWriterMockRecorder) Patch(ctx, obj, patch any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, obj, patch}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Patch", reflect.TypeOf((*MockStatusWriter)(nil).Patch), varargs...)
}

// Update mocks base method.
func (m *MockStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	m.ctrl.T.Helper()
	varargs := []any{ctx, obj}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockStatusWriterMockRecorder) Update(ctx, obj any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, obj}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockStatusWriter)(nil).Update), varargs...)
}
