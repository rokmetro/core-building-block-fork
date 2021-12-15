package storage

import "core-building-block/core/model"

//LoginSession
func loginSessionFromStorage(item loginSession, authType model.AuthType, account *model.Account,
	appOrg model.ApplicationOrganization) model.LoginSession {
	id := item.ID

	appType := model.ApplicationType{ID: item.AppTypeID, Identifier: item.AppTypeIdentifier}

	anonymous := item.Anonymous
	identifier := item.Identifier
	var accountAuthType *model.AccountAuthType
	if item.AccountAuthTypeID != nil {
		accountAuthType = account.GetAccountAuthTypeByID(*item.AccountAuthTypeID)
	}
	device := &model.Device{ID: item.ID}
	idAddress := item.IPAddress
	accessToken := item.AccessToken
	refreshTokens := item.RefreshTokens
	params := item.Params

	var state string
	if item.State != nil {
		state = *item.State
	}
	stateExpires := item.StateExpires
	var mfaAttempts int
	if item.MfaAttempts != nil {
		mfaAttempts = *item.MfaAttempts
	}

	expires := item.Expires
	forceExpires := item.ForceExpires

	dateUpdated := item.DateUpdated
	dateCreated := item.DateCreated

	return model.LoginSession{ID: id, AppOrg: appOrg, AuthType: authType, AppType: appType,
		Anonymous: anonymous, Identifier: identifier, AccountAuthType: accountAuthType,
		Device: device, IPAddress: idAddress, AccessToken: accessToken, RefreshTokens: refreshTokens, Params: params,
		State: state, StateExpires: stateExpires, MfaAttempts: mfaAttempts,
		Expires: expires, ForceExpires: forceExpires, DateUpdated: dateUpdated, DateCreated: dateCreated}
}

func loginSessionToStorage(item model.LoginSession) *loginSession {
	id := item.ID

	appID := item.AppOrg.Application.ID
	orgID := item.AppOrg.Organization.ID

	authTypeCode := item.AuthType.Code

	appTypeID := item.AppType.ID
	appTypeIdentifier := item.AppType.Identifier

	anonymous := item.Anonymous
	identifier := item.Identifier
	var accountAuthTypeID *string
	var accountAuthTypeIdentifier *string
	if item.AccountAuthType != nil {
		accountAuthTypeID = &item.AccountAuthType.ID
		accountAuthTypeIdentifier = &item.AccountAuthType.Identifier
	}
	var deviceID *string
	if item.Device != nil {
		deviceID = &item.Device.ID
	}
	ipAddress := item.IPAddress
	accessToken := item.AccessToken
	refreshTokens := item.RefreshTokens
	params := item.Params

	var state *string
	if item.State != "" {
		state = &item.State
	}
	stateExpires := item.StateExpires
	var mfaAttempts *int
	if item.MfaAttempts != 0 {
		mfaAttempts = &item.MfaAttempts
	}

	expires := item.Expires
	forceExpires := item.ForceExpires

	dateUpdated := item.DateUpdated
	dateCreated := item.DateCreated

	return &loginSession{ID: id, AppID: appID, OrgID: orgID, AuthTypeCode: authTypeCode,
		AppTypeID: appTypeID, AppTypeIdentifier: appTypeIdentifier, Anonymous: anonymous,
		Identifier: identifier, AccountAuthTypeID: accountAuthTypeID, AccountAuthTypeIdentifier: accountAuthTypeIdentifier,
		DeviceID: deviceID, IPAddress: ipAddress, AccessToken: accessToken, RefreshTokens: refreshTokens,
		Params: params, State: state, StateExpires: stateExpires, MfaAttempts: mfaAttempts,
		Expires: expires, ForceExpires: forceExpires, DateUpdated: dateUpdated, DateCreated: dateCreated}
}

func logSessionFromStorage(items *model.LoginSession) model.LoginSession {

	return model.LoginSession{ID: items.ID, AppOrg: items.AppOrg, AuthType: items.AuthType, AppType: items.AppType,
		Anonymous: items.Anonymous, Identifier: items.Identifier, AccountAuthType: items.AccountAuthType,
		Device: items.Device, IPAddress: items.IPAddress, AccessToken: items.AccessToken, RefreshTokens: items.RefreshTokens, Params: items.Params,
		State: items.State, StateExpires: items.StateExpires, MfaAttempts: items.MfaAttempts,
		Expires: items.Expires, ForceExpires: items.ForceExpires, DateUpdated: items.DateUpdated, DateCreated: items.DateCreated}
}

func logSessionsFromStorage(itemsList []model.LoginSession) []model.LoginSession {
	if len(itemsList) == 0 {
		return make([]model.LoginSession, 0)
	}

	var items []model.LoginSession
	for _, ls := range itemsList {
		items = append(items, logSessionFromStorage(&ls))
	}
	return items
}
