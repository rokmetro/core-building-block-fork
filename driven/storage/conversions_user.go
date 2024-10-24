// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"core-building-block/core/model"
)

// Account
func accountFromStorage(item account, appOrg model.ApplicationOrganization, sa *Adapter) model.Account {
	roles := accountRolesFromStorage(item.Roles, appOrg)
	groups := accountGroupsFromStorage(item.Groups, appOrg)
	identifiers := accountIdentifiersFromStorage(item.Identifiers)
	authTypes := accountAuthTypesFromStorage(item.AuthTypes, sa)
	mfaTypes := mfaTypesFromStorage(item.MFATypes)
	profile := profileFromStorage(item.Profile)
	devices := accountDevicesFromStorage(item)
	return model.Account{ID: item.ID, AppOrg: appOrg, Anonymous: item.Anonymous, Permissions: item.Permissions, Roles: roles, Groups: groups, Scopes: item.Scopes,
		Identifiers: identifiers, AuthTypes: authTypes, MFATypes: mfaTypes, Preferences: item.Preferences, Profile: profile, SystemConfigs: item.SystemConfigs,
		Secrets: item.Secrets, Privacy: item.Privacy, Verified: item.Verified, Devices: devices, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated,
		LastLoginDate: item.LastLoginDate, LastAccessTokenDate: item.LastAccessTokenDate, MostRecentClientVersion: item.MostRecentClientVersion}
}

func accountsFromStorage(items []account, appOrg model.ApplicationOrganization, sa *Adapter) []model.Account {
	if len(items) == 0 {
		return make([]model.Account, 0)
	}

	res := make([]model.Account, len(items))
	for i, item := range items {
		res[i] = accountFromStorage(item, appOrg, sa)
	}
	return res
}

func accountToStorage(item *model.Account) *account {
	id := item.ID
	appOrgID := item.AppOrg.ID
	permissions := item.Permissions
	roles := accountRolesToStorage(item.Roles)
	groups := accountGroupsToStorage(item.Groups)
	identifiers := accountIdentifiersToStorage(item.Identifiers)
	authTypes := accountAuthTypesToStorage(item.AuthTypes)
	mfaTypes := mfaTypesToStorage(item.MFATypes)
	profile := profileToStorage(item.Profile)
	devices := accountDevicesToStorage(item)
	dateCreated := item.DateCreated
	dateUpdated := item.DateUpdated
	lastLoginDate := item.LastLoginDate
	lastAccessTokenDate := item.LastAccessTokenDate
	mostRecentClientVersion := item.MostRecentClientVersion

	return &account{ID: id, AppOrgID: appOrgID, Anonymous: item.Anonymous, Permissions: permissions, Roles: roles, Groups: groups, Scopes: item.Scopes,
		Identifiers: identifiers, AuthTypes: authTypes, MFATypes: mfaTypes, Privacy: item.Privacy, Verified: item.Verified, Preferences: item.Preferences,
		Profile: profile, Secrets: item.Secrets, SystemConfigs: item.SystemConfigs, Devices: devices, DateCreated: dateCreated, DateUpdated: dateUpdated,
		LastLoginDate: lastLoginDate, LastAccessTokenDate: lastAccessTokenDate, MostRecentClientVersion: mostRecentClientVersion}
}

func accountDevicesFromStorage(item account) []model.Device {
	devices := make([]model.Device, len(item.Devices))

	for i, device := range item.Devices {
		devices[i] = accountDeviceFromStorage(device)
	}
	return devices
}

func accountDeviceFromStorage(item userDevice) model.Device {
	return model.Device{ID: item.ID, DeviceID: item.DeviceID, Type: item.Type, OS: item.OS,
		DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func accountDevicesToStorage(item *model.Account) []userDevice {
	devices := make([]userDevice, len(item.Devices))

	for i, device := range item.Devices {
		devices[i] = accountDeviceToStorage(device)
	}
	return devices
}

func accountDeviceToStorage(item model.Device) userDevice {
	return userDevice{ID: item.ID, DeviceID: item.DeviceID, Type: item.Type, OS: item.OS,
		DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

// AccountAuthType
func accountAuthTypeFromStorage(item accountAuthType, sa *Adapter) model.AccountAuthType {
	id := item.ID

	authType, _ := sa.FindAuthType(item.AuthTypeID)
	params := item.Params
	var credential *model.Credential
	if item.CredentialID != nil {
		credential = &model.Credential{ID: *item.CredentialID}
	}
	active := item.Active
	return model.AccountAuthType{ID: id, SupportedAuthType: model.SupportedAuthType{AuthTypeID: item.AuthTypeID, AuthType: *authType}, Params: params, Credential: credential,
		Active: active, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func accountAuthTypesFromStorage(items []accountAuthType, sa *Adapter) []model.AccountAuthType {
	res := make([]model.AccountAuthType, len(items))
	for i, aat := range items {
		res[i] = accountAuthTypeFromStorage(aat, sa)
	}
	return res
}

func accountAuthTypeToStorage(item model.AccountAuthType) accountAuthType {
	var credentialID *string
	if item.Credential != nil {
		credentialID = &item.Credential.ID
	}
	return accountAuthType{ID: item.ID, AuthTypeID: item.SupportedAuthType.AuthType.ID, AuthTypeCode: item.SupportedAuthType.AuthType.Code,
		Params: item.Params, CredentialID: credentialID, Active: item.Active, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func accountAuthTypesToStorage(items []model.AccountAuthType) []accountAuthType {
	res := make([]accountAuthType, len(items))
	for i, aat := range items {
		res[i] = accountAuthTypeToStorage(aat)
	}
	return res
}

// AccountIdentifier
func accountIdentifierFromStorage(item accountIdentifier) model.AccountIdentifier {
	return model.AccountIdentifier{ID: item.ID, Code: item.Code, Identifier: item.Identifier, Verified: item.Verified, Linked: item.Linked,
		Sensitive: item.Sensitive, AccountAuthTypeID: item.AccountAuthTypeID, Primary: item.Primary, VerificationCode: item.VerificationCode,
		VerificationExpiry: item.VerificationExpiry, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func accountIdentifiersFromStorage(items []accountIdentifier) []model.AccountIdentifier {
	res := make([]model.AccountIdentifier, len(items))
	for i, aat := range items {
		res[i] = accountIdentifierFromStorage(aat)
	}
	return res
}

func accountIdentifierToStorage(item model.AccountIdentifier) accountIdentifier {
	return accountIdentifier{ID: item.ID, Code: item.Code, Identifier: item.Identifier, Verified: item.Verified, Linked: item.Linked,
		Sensitive: item.Sensitive, AccountAuthTypeID: item.AccountAuthTypeID, Primary: item.Primary, VerificationCode: item.VerificationCode,
		VerificationExpiry: item.VerificationExpiry, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func accountIdentifiersToStorage(items []model.AccountIdentifier) []accountIdentifier {
	res := make([]accountIdentifier, len(items))
	for i, aat := range items {
		res[i] = accountIdentifierToStorage(aat)
	}
	return res
}

// AccountRole
func accountRoleFromStorage(item *accountRole, appOrg model.ApplicationOrganization) model.AccountRole {
	if item == nil {
		return model.AccountRole{}
	}

	appOrgRole := appOrgRoleFromStorage(&item.Role, appOrg)
	return model.AccountRole{Role: appOrgRole, Active: item.Active, AdminSet: item.AdminSet}
}

func accountRolesFromStorage(items []accountRole, application model.ApplicationOrganization) []model.AccountRole {
	if len(items) == 0 {
		return make([]model.AccountRole, 0)
	}

	res := make([]model.AccountRole, len(items))
	for i, item := range items {
		res[i] = accountRoleFromStorage(&item, application)
	}
	return res
}

func accountRoleToStorage(item model.AccountRole) accountRole {
	appRole := appOrgRoleToStorage(item.Role)
	return accountRole{Role: appRole, Active: item.Active, AdminSet: item.AdminSet}
}

func accountRolesToStorage(items []model.AccountRole) []accountRole {
	if len(items) == 0 {
		return make([]accountRole, 0)
	}

	res := make([]accountRole, len(items))
	for i, item := range items {
		res[i] = accountRoleToStorage(item)
	}
	return res
}

// ApplicationGroup
func accountGroupFromStorage(item *accountGroup, appOrg model.ApplicationOrganization) model.AccountGroup {
	if item == nil {
		return model.AccountGroup{}
	}

	appOrgGroup := appOrgGroupFromStorage(&item.Group, appOrg)
	return model.AccountGroup{Group: appOrgGroup, Active: item.Active, AdminSet: item.AdminSet}
}

func accountGroupsFromStorage(items []accountGroup, appOrg model.ApplicationOrganization) []model.AccountGroup {
	if len(items) == 0 {
		return make([]model.AccountGroup, 0)
	}

	res := make([]model.AccountGroup, len(items))
	for i, item := range items {
		res[i] = accountGroupFromStorage(&item, appOrg)
	}
	return res
}

func accountGroupToStorage(item model.AccountGroup) accountGroup {
	appGroup := appOrgGroupToStorage(item.Group)
	return accountGroup{Group: appGroup, Active: item.Active, AdminSet: item.AdminSet}
}

func accountGroupsToStorage(items []model.AccountGroup) []accountGroup {
	if len(items) == 0 {
		return make([]accountGroup, 0)
	}

	res := make([]accountGroup, len(items))
	for i, item := range items {
		res[i] = accountGroupToStorage(item)
	}
	return res
}

// Profile
func profileFromStorage(item profile) model.Profile {
	return model.Profile{ID: item.ID, PhotoURL: item.PhotoURL, FirstName: item.FirstName, LastName: item.LastName,
		BirthYear: item.BirthYear, Address: item.Address, ZipCode: item.ZipCode, State: item.State,
		Country: item.Country, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated,
		UnstructuredProperties: item.UnstructuredProperties}
}

func profilesFromStorage(items []account, sa *Adapter) []model.Profile {
	if len(items) == 0 {
		return make([]model.Profile, 0)
	}

	//prepare accounts
	accounts := make(map[string][]model.Account, len(items))
	for _, account := range items {
		appOrg, _ := sa.getCachedApplicationOrganizationByKey(account.AppOrgID)
		rAccount := accountFromStorage(account, *appOrg, sa)

		//add account to the map
		profileAccounts := accounts[rAccount.Profile.ID]
		if profileAccounts == nil {
			profileAccounts = []model.Account{}
		}
		profileAccounts = append(profileAccounts, rAccount)
		accounts[rAccount.Profile.ID] = profileAccounts
	}

	//prepare profiles
	res := make([]model.Profile, len(items))
	for i, item := range items {

		profile := profileFromStorage(item.Profile)
		profile.Accounts = accounts[item.Profile.ID]

		res[i] = profile
	}
	return res
}

func profileToStorage(item model.Profile) profile {
	return profile{ID: item.ID, PhotoURL: item.PhotoURL, FirstName: item.FirstName, LastName: item.LastName,
		BirthYear: item.BirthYear, Address: item.Address, ZipCode: item.ZipCode, State: item.State,
		Country: item.Country, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated,
		UnstructuredProperties: item.UnstructuredProperties}
}

// Device
func deviceToStorage(item *model.Device) *device {
	if item == nil {
		return nil
	}

	return &device{ID: item.ID, DeviceID: item.DeviceID, Type: item.Type, OS: item.OS,
		Account: item.Account.ID, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func deviceFromStorage(item device) model.Device {
	return model.Device{ID: item.ID, DeviceID: item.DeviceID, Type: item.Type, OS: item.OS, DateUpdated: item.DateUpdated}
}

// Credential
func credentialFromStorage(item credential) model.Credential {
	accountAuthTypes := make([]model.AccountAuthType, len(item.AccountsAuthTypes))
	for i, id := range item.AccountsAuthTypes {
		accountAuthTypes[i] = model.AccountAuthType{ID: id}
	}
	authType := model.AuthType{ID: item.AuthTypeID}
	return model.Credential{ID: item.ID, AuthType: authType, AccountsAuthTypes: accountAuthTypes,
		Value: item.Value, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func credentialsFromStorage(items []credential) []model.Credential {
	res := make([]model.Credential, len(items))
	for i, cred := range items {
		res[i] = credentialFromStorage(cred)
	}
	return res
}

func credentialToStorage(item model.Credential) credential {
	accountAuthTypes := make([]string, len(item.AccountsAuthTypes))
	for i, aat := range item.AccountsAuthTypes {
		accountAuthTypes[i] = aat.ID
	}
	return credential{ID: item.ID, AuthTypeID: item.AuthType.ID, AccountsAuthTypes: accountAuthTypes,
		Value: item.Value, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}
}

func credentialsToStorage(items []model.Credential) []credential {
	res := make([]credential, len(items))
	for i, cred := range items {
		res[i] = credentialToStorage(cred)
	}
	return res
}

// MFA
func mfaTypesFromStorage(items []mfaType) []model.MFAType {
	res := make([]model.MFAType, len(items))
	for i, mfa := range items {
		res[i] = mfaTypeFromStorage(mfa)
	}
	return res
}

func mfaTypeFromStorage(item mfaType) model.MFAType {
	return model.MFAType{ID: item.ID, Type: item.Type, Verified: item.Verified, Params: item.Params, DateCreated: item.DateCreated,
		DateUpdated: item.DateUpdated}
}

func mfaTypesToStorage(items []model.MFAType) []mfaType {
	res := make([]mfaType, len(items))
	for i, mfa := range items {
		res[i] = mfaTypeToStorage(&mfa)
	}
	return res
}

func mfaTypeToStorage(item *model.MFAType) mfaType {
	//don't store totp qr code
	params := make(map[string]interface{})
	for k, v := range item.Params {
		if k != "qr_code" {
			params[k] = v
		}
	}

	return mfaType{ID: item.ID, Type: item.Type, Verified: item.Verified, Params: params, DateCreated: item.DateCreated,
		DateUpdated: item.DateUpdated}
}
