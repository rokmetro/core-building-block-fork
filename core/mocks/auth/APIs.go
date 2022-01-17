// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	authorization "github.com/rokwire/core-auth-library-go/authorization"
	logs "github.com/rokwire/logging-library-go/logs"

	mock "github.com/stretchr/testify/mock"

	model "core-building-block/core/model"

	tokenauth "github.com/rokwire/core-auth-library-go/tokenauth"
)

// APIs is an autogenerated mock type for the APIs type
type APIs struct {
	mock.Mock
}

// AccountExists provides a mock function with given fields: authenticationType, userIdentifier, apiKey, appTypeIdentifier, orgID, l
func (_m *APIs) AccountExists(authenticationType string, userIdentifier string, apiKey string, appTypeIdentifier string, orgID string, l *logs.Log) (bool, error) {
	ret := _m.Called(authenticationType, userIdentifier, apiKey, appTypeIdentifier, orgID, l)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *logs.Log) bool); ok {
		r0 = rf(authenticationType, userIdentifier, apiKey, appTypeIdentifier, orgID, l)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, *logs.Log) error); ok {
		r1 = rf(authenticationType, userIdentifier, apiKey, appTypeIdentifier, orgID, l)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AuthorizeService provides a mock function with given fields: claims, serviceID, approvedScopes, l
func (_m *APIs) AuthorizeService(claims tokenauth.Claims, serviceID string, approvedScopes []authorization.Scope, l *logs.Log) (string, []authorization.Scope, *model.ServiceReg, error) {
	ret := _m.Called(claims, serviceID, approvedScopes, l)

	var r0 string
	if rf, ok := ret.Get(0).(func(tokenauth.Claims, string, []authorization.Scope, *logs.Log) string); ok {
		r0 = rf(claims, serviceID, approvedScopes, l)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 []authorization.Scope
	if rf, ok := ret.Get(1).(func(tokenauth.Claims, string, []authorization.Scope, *logs.Log) []authorization.Scope); ok {
		r1 = rf(claims, serviceID, approvedScopes, l)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]authorization.Scope)
		}
	}

	var r2 *model.ServiceReg
	if rf, ok := ret.Get(2).(func(tokenauth.Claims, string, []authorization.Scope, *logs.Log) *model.ServiceReg); ok {
		r2 = rf(claims, serviceID, approvedScopes, l)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*model.ServiceReg)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(tokenauth.Claims, string, []authorization.Scope, *logs.Log) error); ok {
		r3 = rf(claims, serviceID, approvedScopes, l)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// CreateAPIKey provides a mock function with given fields: apiKey
func (_m *APIs) CreateAPIKey(apiKey model.APIKey) (*model.APIKey, error) {
	ret := _m.Called(apiKey)

	var r0 *model.APIKey
	if rf, ok := ret.Get(0).(func(model.APIKey) *model.APIKey); ok {
		r0 = rf(apiKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIKey)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.APIKey) error); ok {
		r1 = rf(apiKey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAPIKey provides a mock function with given fields: ID
func (_m *APIs) DeleteAPIKey(ID string) error {
	ret := _m.Called(ID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(ID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeregisterService provides a mock function with given fields: serviceID
func (_m *APIs) DeregisterService(serviceID string) error {
	ret := _m.Called(serviceID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(serviceID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ForgotCredential provides a mock function with given fields: authenticationType, appTypeIdentifier, orgID, apiKey, identifier, l
func (_m *APIs) ForgotCredential(authenticationType string, appTypeIdentifier string, orgID string, apiKey string, identifier string, l *logs.Log) error {
	ret := _m.Called(authenticationType, appTypeIdentifier, orgID, apiKey, identifier, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *logs.Log) error); ok {
		r0 = rf(authenticationType, appTypeIdentifier, orgID, apiKey, identifier, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAPIKey provides a mock function with given fields: ID
func (_m *APIs) GetAPIKey(ID string) (*model.APIKey, error) {
	ret := _m.Called(ID)

	var r0 *model.APIKey
	if rf, ok := ret.Get(0).(func(string) *model.APIKey); ok {
		r0 = rf(ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIKey)
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

// GetApplicationAPIKeys provides a mock function with given fields: appID
func (_m *APIs) GetApplicationAPIKeys(appID string) ([]model.APIKey, error) {
	ret := _m.Called(appID)

	var r0 []model.APIKey
	if rf, ok := ret.Get(0).(func(string) []model.APIKey); ok {
		r0 = rf(appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.APIKey)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAuthKeySet provides a mock function with given fields:
func (_m *APIs) GetAuthKeySet() (*model.JSONWebKeySet, error) {
	ret := _m.Called()

	var r0 *model.JSONWebKeySet
	if rf, ok := ret.Get(0).(func() *model.JSONWebKeySet); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.JSONWebKeySet)
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

// GetHost provides a mock function with given fields:
func (_m *APIs) GetHost() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetLoginURL provides a mock function with given fields: authType, appTypeIdentifier, orgID, redirectURI, apiKey, l
func (_m *APIs) GetLoginURL(authType string, appTypeIdentifier string, orgID string, redirectURI string, apiKey string, l *logs.Log) (string, map[string]interface{}, error) {
	ret := _m.Called(authType, appTypeIdentifier, orgID, redirectURI, apiKey, l)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *logs.Log) string); ok {
		r0 = rf(authType, appTypeIdentifier, orgID, redirectURI, apiKey, l)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 map[string]interface{}
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, *logs.Log) map[string]interface{}); ok {
		r1 = rf(authType, appTypeIdentifier, orgID, redirectURI, apiKey, l)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[string]interface{})
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string, string, string, string, *logs.Log) error); ok {
		r2 = rf(authType, appTypeIdentifier, orgID, redirectURI, apiKey, l)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetScopedAccessToken provides a mock function with given fields: claims, serviceID, scopes
func (_m *APIs) GetScopedAccessToken(claims tokenauth.Claims, serviceID string, scopes []authorization.Scope) (string, error) {
	ret := _m.Called(claims, serviceID, scopes)

	var r0 string
	if rf, ok := ret.Get(0).(func(tokenauth.Claims, string, []authorization.Scope) string); ok {
		r0 = rf(claims, serviceID, scopes)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(tokenauth.Claims, string, []authorization.Scope) error); ok {
		r1 = rf(claims, serviceID, scopes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetServiceRegistrations provides a mock function with given fields: serviceIDs
func (_m *APIs) GetServiceRegistrations(serviceIDs []string) ([]model.ServiceReg, error) {
	ret := _m.Called(serviceIDs)

	var r0 []model.ServiceReg
	if rf, ok := ret.Get(0).(func([]string) []model.ServiceReg); ok {
		r0 = rf(serviceIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.ServiceReg)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]string) error); ok {
		r1 = rf(serviceIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LinkAccountAuthType provides a mock function with given fields: accountID, authenticationType, appTypeIdentifier, creds, params, l
func (_m *APIs) LinkAccountAuthType(accountID string, authenticationType string, appTypeIdentifier string, creds string, params string, l *logs.Log) (*string, error) {
	ret := _m.Called(accountID, authenticationType, appTypeIdentifier, creds, params, l)

	var r0 *string
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *logs.Log) *string); ok {
		r0 = rf(accountID, authenticationType, appTypeIdentifier, creds, params, l)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, *logs.Log) error); ok {
		r1 = rf(accountID, authenticationType, appTypeIdentifier, creds, params, l)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Login provides a mock function with given fields: ipAddress, deviceType, deviceOS, deviceID, authenticationType, creds, apiKey, appTypeIdentifier, orgID, params, profile, preferences, l
func (_m *APIs) Login(ipAddress string, deviceType string, deviceOS *string, deviceID string, authenticationType string, creds string, apiKey string, appTypeIdentifier string, orgID string, params string, profile model.Profile, preferences map[string]interface{}, l *logs.Log) (*string, *model.LoginSession, error) {
	ret := _m.Called(ipAddress, deviceType, deviceOS, deviceID, authenticationType, creds, apiKey, appTypeIdentifier, orgID, params, profile, preferences, l)

	var r0 *string
	if rf, ok := ret.Get(0).(func(string, string, *string, string, string, string, string, string, string, string, model.Profile, map[string]interface{}, *logs.Log) *string); ok {
		r0 = rf(ipAddress, deviceType, deviceOS, deviceID, authenticationType, creds, apiKey, appTypeIdentifier, orgID, params, profile, preferences, l)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	var r1 *model.LoginSession
	if rf, ok := ret.Get(1).(func(string, string, *string, string, string, string, string, string, string, string, model.Profile, map[string]interface{}, *logs.Log) *model.LoginSession); ok {
		r1 = rf(ipAddress, deviceType, deviceOS, deviceID, authenticationType, creds, apiKey, appTypeIdentifier, orgID, params, profile, preferences, l)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.LoginSession)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string, *string, string, string, string, string, string, string, string, model.Profile, map[string]interface{}, *logs.Log) error); ok {
		r2 = rf(ipAddress, deviceType, deviceOS, deviceID, authenticationType, creds, apiKey, appTypeIdentifier, orgID, params, profile, preferences, l)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Refresh provides a mock function with given fields: refreshToken, apiKey, l
func (_m *APIs) Refresh(refreshToken string, apiKey string, l *logs.Log) (*model.LoginSession, error) {
	ret := _m.Called(refreshToken, apiKey, l)

	var r0 *model.LoginSession
	if rf, ok := ret.Get(0).(func(string, string, *logs.Log) *model.LoginSession); ok {
		r0 = rf(refreshToken, apiKey, l)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.LoginSession)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *logs.Log) error); ok {
		r1 = rf(refreshToken, apiKey, l)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterService provides a mock function with given fields: reg
func (_m *APIs) RegisterService(reg *model.ServiceReg) error {
	ret := _m.Called(reg)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.ServiceReg) error); ok {
		r0 = rf(reg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResetForgotCredential provides a mock function with given fields: credsID, resetCode, params, l
func (_m *APIs) ResetForgotCredential(credsID string, resetCode string, params string, l *logs.Log) error {
	ret := _m.Called(credsID, resetCode, params, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, *logs.Log) error); ok {
		r0 = rf(credsID, resetCode, params, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendVerifyCredential provides a mock function with given fields: authenticationType, appTypeIdentifier, orgID, apiKey, identifier, l
func (_m *APIs) SendVerifyCredential(authenticationType string, appTypeIdentifier string, orgID string, apiKey string, identifier string, l *logs.Log) error {
	ret := _m.Called(authenticationType, appTypeIdentifier, orgID, apiKey, identifier, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, *logs.Log) error); ok {
		r0 = rf(authenticationType, appTypeIdentifier, orgID, apiKey, identifier, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields:
func (_m *APIs) Start() {
	_m.Called()
}

// UpdateAPIKey provides a mock function with given fields: apiKey
func (_m *APIs) UpdateAPIKey(apiKey model.APIKey) error {
	ret := _m.Called(apiKey)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.APIKey) error); ok {
		r0 = rf(apiKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateCredential provides a mock function with given fields: accountID, accountAuthTypeID, params, l
func (_m *APIs) UpdateCredential(accountID string, accountAuthTypeID string, params string, l *logs.Log) error {
	ret := _m.Called(accountID, accountAuthTypeID, params, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, *logs.Log) error); ok {
		r0 = rf(accountID, accountAuthTypeID, params, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateServiceRegistration provides a mock function with given fields: reg
func (_m *APIs) UpdateServiceRegistration(reg *model.ServiceReg) error {
	ret := _m.Called(reg)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.ServiceReg) error); ok {
		r0 = rf(reg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// VerifyCredential provides a mock function with given fields: id, verification, l
func (_m *APIs) VerifyCredential(id string, verification string, l *logs.Log) error {
	ret := _m.Called(id, verification, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, *logs.Log) error); ok {
		r0 = rf(id, verification, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}