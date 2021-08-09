// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	model "core-building-block/core/model"

	mock "github.com/stretchr/testify/mock"

	storage "core-building-block/driven/storage"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

// CreateGlobalConfig provides a mock function with given fields: setting
func (_m *Storage) CreateGlobalConfig(setting string) (*model.GlobalConfig, error) {
	ret := _m.Called(setting)

	var r0 *model.GlobalConfig
	if rf, ok := ret.Get(0).(func(string) *model.GlobalConfig); ok {
		r0 = rf(setting)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.GlobalConfig)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(setting)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateOrganization provides a mock function with given fields: name, requestType, requiresOwnLogin, loginTypes, organizationDomains
func (_m *Storage) CreateOrganization(name string, requestType string, requiresOwnLogin bool, loginTypes []string, organizationDomains []string) (*model.Organization, error) {
	ret := _m.Called(name, requestType, requiresOwnLogin, loginTypes, organizationDomains)

	var r0 *model.Organization
	if rf, ok := ret.Get(0).(func(string, string, bool, []string, []string) *model.Organization); ok {
		r0 = rf(name, requestType, requiresOwnLogin, loginTypes, organizationDomains)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Organization)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, bool, []string, []string) error); ok {
		r1 = rf(name, requestType, requiresOwnLogin, loginTypes, organizationDomains)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetApplication provides a mock function with given fields: ID
func (_m *Storage) GetApplication(ID string) (*model.Application, error) {
	ret := _m.Called(ID)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(string) *model.Application); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobalConfig provides a mock function with given fields:
func (_m *Storage) GetGlobalConfig() (*model.GlobalConfig, error) {
	ret := _m.Called()

	var r0 *model.GlobalConfig
	if rf, ok := ret.Get(0).(func() *model.GlobalConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.GlobalConfig)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobalPermissions provides a mock function with given fields:
func (_m *Storage) GetGlobalPermissions() ([]model.GlobalPermission, error) {
	ret := _m.Called()

	var r0 []model.GlobalPermission
	if rf, ok := ret.Get(0).(func() []model.GlobalPermission); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.GlobalPermission)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOrganization provides a mock function with given fields: ID
func (_m *Storage) GetOrganization(ID string) (*model.Organization, error) {
	ret := _m.Called(ID)

	var r0 *model.Organization
	if rf, ok := ret.Get(0).(func(string) *model.Organization); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Organization)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetOrganizations provides a mock function with given fields:
func (_m *Storage) GetOrganizations() ([]model.Organization, error) {
	ret := _m.Called()

	var r0 []model.Organization
	if rf, ok := ret.Get(0).(func() []model.Organization); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Organization)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterStorageListener provides a mock function with given fields: storageListener
func (_m *Storage) RegisterStorageListener(storageListener storage.Listener) {
	_m.Called(storageListener)
}

// SaveGlobalConfig provides a mock function with given fields: setting
func (_m *Storage) SaveGlobalConfig(setting *model.GlobalConfig) error {
	ret := _m.Called(setting)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.GlobalConfig) error); ok {
		r0 = rf(setting)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateOrganization provides a mock function with given fields: ID, name, requestType, requiresOwnLogin, loginTypes, organizationDomains
func (_m *Storage) UpdateOrganization(ID string, name string, requestType string, requiresOwnLogin bool, loginTypes []string, organizationDomains []string) error {
	ret := _m.Called(ID, name, requestType, requiresOwnLogin, loginTypes, organizationDomains)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, bool, []string, []string) error); ok {
		r0 = rf(ID, name, requestType, requiresOwnLogin, loginTypes, organizationDomains)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
