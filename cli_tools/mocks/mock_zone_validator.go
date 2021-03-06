//  Copyright 2019 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/GoogleCloudPlatform/compute-image-import/cli_tools/common/domain (interfaces: ZoneValidatorInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockZoneValidatorInterface is a mock of ZoneValidatorInterface interface
type MockZoneValidatorInterface struct {
	ctrl     *gomock.Controller
	recorder *MockZoneValidatorInterfaceMockRecorder
}

// MockZoneValidatorInterfaceMockRecorder is the mock recorder for MockZoneValidatorInterface
type MockZoneValidatorInterfaceMockRecorder struct {
	mock *MockZoneValidatorInterface
}

// NewMockZoneValidatorInterface creates a new mock instance
func NewMockZoneValidatorInterface(ctrl *gomock.Controller) *MockZoneValidatorInterface {
	mock := &MockZoneValidatorInterface{ctrl: ctrl}
	mock.recorder = &MockZoneValidatorInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockZoneValidatorInterface) EXPECT() *MockZoneValidatorInterfaceMockRecorder {
	return m.recorder
}

// ZoneValid mocks base method
func (m *MockZoneValidatorInterface) ZoneValid(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ZoneValid", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ZoneValid indicates an expected call of ZoneValid
func (mr *MockZoneValidatorInterfaceMockRecorder) ZoneValid(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ZoneValid", reflect.TypeOf((*MockZoneValidatorInterface)(nil).ZoneValid), arg0, arg1)
}
