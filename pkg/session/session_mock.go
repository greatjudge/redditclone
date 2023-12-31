// Code generated by MockGen. DO NOT EDIT.
// Source: session.go

// Package session is a generated GoMock package.
package session

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	user "github.com/greatjudge/redditclone/pkg/user"
)

// MockSessionsManager is a mock of SessionsManager interface.
type MockSessionsManager struct {
	ctrl     *gomock.Controller
	recorder *MockSessionsManagerMockRecorder
}

// MockSessionsManagerMockRecorder is the mock recorder for MockSessionsManager.
type MockSessionsManagerMockRecorder struct {
	mock *MockSessionsManager
}

// NewMockSessionsManager creates a new mock instance.
func NewMockSessionsManager(ctrl *gomock.Controller) *MockSessionsManager {
	mock := &MockSessionsManager{ctrl: ctrl}
	mock.recorder = &MockSessionsManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionsManager) EXPECT() *MockSessionsManagerMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockSessionsManager) Check(r *http.Request) (Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", r)
	ret0, _ := ret[0].(Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Check indicates an expected call of Check.
func (mr *MockSessionsManagerMockRecorder) Check(r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockSessionsManager)(nil).Check), r)
}

// Create mocks base method.
func (m *MockSessionsManager) Create(user user.User) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", user)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockSessionsManagerMockRecorder) Create(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSessionsManager)(nil).Create), user)
}
