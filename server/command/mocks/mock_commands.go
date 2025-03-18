// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-shoutrrr/server/command (interfaces: Command)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/mattermost/mattermost/server/public/model"
)

// MockCommand is a mock of Command interface.
type MockCommand struct {
	ctrl     *gomock.Controller
	recorder *MockCommandMockRecorder
}

// MockCommandMockRecorder is the mock recorder for MockCommand.
type MockCommandMockRecorder struct {
	mock *MockCommand
}

// NewMockCommand creates a new mock instance.
func NewMockCommand(ctrl *gomock.Controller) *MockCommand {
	mock := &MockCommand{ctrl: ctrl}
	mock.recorder = &MockCommandMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCommand) EXPECT() *MockCommandMockRecorder {
	return m.recorder
}

// Handle mocks base method.
func (m *MockCommand) Handle(arg0 *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", arg0)
	ret0, _ := ret[0].(*model.CommandResponse)
	ret1, _ := ret[1].(*model.AppError)
	return ret0, ret1
}

// Handle indicates an expected call of Handle.
func (mr *MockCommandMockRecorder) Handle(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockCommand)(nil).Handle), arg0)
}

// executeHelloCommand mocks base method.
func (m *MockCommand) executeHelloCommand(arg0 *model.CommandArgs) *model.CommandResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "executeHelloCommand", arg0)
	ret0, _ := ret[0].(*model.CommandResponse)
	return ret0
}

// executeHelloCommand indicates an expected call of executeHelloCommand.
func (mr *MockCommandMockRecorder) executeHelloCommand(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "executeHelloCommand", reflect.TypeOf((*MockCommand)(nil).executeHelloCommand), arg0)
}
