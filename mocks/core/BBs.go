// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// BBs is an autogenerated mock type for the BBs type
type BBs struct {
	mock.Mock
}

// BBsGetTest provides a mock function with given fields:
func (_m *BBs) BBsGetTest() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}