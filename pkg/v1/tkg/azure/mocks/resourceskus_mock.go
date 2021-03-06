// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute/computeapi (interfaces: ResourceSkusClientAPI)

// Package azure is a generated GoMock package.
package azure

import (
	context "context"
	reflect "reflect"

	compute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	gomock "github.com/golang/mock/gomock"
)

// MockResourceSkusClientAPI is a mock of ResourceSkusClientAPI interface.
type MockResourceSkusClientAPI struct {
	ctrl     *gomock.Controller
	recorder *MockResourceSkusClientAPIMockRecorder
}

// MockResourceSkusClientAPIMockRecorder is the mock recorder for MockResourceSkusClientAPI.
type MockResourceSkusClientAPIMockRecorder struct {
	mock *MockResourceSkusClientAPI
}

// NewMockResourceSkusClientAPI creates a new mock instance.
func NewMockResourceSkusClientAPI(ctrl *gomock.Controller) *MockResourceSkusClientAPI {
	mock := &MockResourceSkusClientAPI{ctrl: ctrl}
	mock.recorder = &MockResourceSkusClientAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceSkusClientAPI) EXPECT() *MockResourceSkusClientAPIMockRecorder {
	return m.recorder
}

// List mocks base method.
func (m *MockResourceSkusClientAPI) List(arg0 context.Context, arg1 string) (compute.ResourceSkusResultPage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", arg0, arg1)
	ret0, _ := ret[0].(compute.ResourceSkusResultPage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockResourceSkusClientAPIMockRecorder) List(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockResourceSkusClientAPI)(nil).List), arg0, arg1)
}

// ListComplete mocks base method.
func (m *MockResourceSkusClientAPI) ListComplete(arg0 context.Context, arg1 string) (compute.ResourceSkusResultIterator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListComplete", arg0, arg1)
	ret0, _ := ret[0].(compute.ResourceSkusResultIterator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListComplete indicates an expected call of ListComplete.
func (mr *MockResourceSkusClientAPIMockRecorder) ListComplete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListComplete", reflect.TypeOf((*MockResourceSkusClientAPI)(nil).ListComplete), arg0, arg1)
}
