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

package web

import (
	"core-building-block/core"
	"core-building-block/core/model"
	Def "core-building-block/driver/web/docs/gen"
	"core-building-block/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authorization"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ServicesApisHandler handles the rest APIs implementation
type ServicesApisHandler struct {
	coreAPIs *core.APIs
}

func (h ServicesApisHandler) login(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	ip := utils.GetIP(l, r)
	clientVersion := r.Header.Get("CLIENT_VERSION")

	var requestData Def.SharedReqLogin
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("auth login request"), nil, err, http.StatusBadRequest, true)
	}

	//creds
	requestCreds, err := interfaceToJSON(requestData.Creds)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	//params
	requestParams, err := interfaceToJSON(requestData.Params)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "params", nil, err, http.StatusBadRequest, true)
	}

	//preferences
	var requestPreferences map[string]interface{}
	if requestData.Preferences != nil {
		requestPreferences = *requestData.Preferences
	}

	//profile ////
	requestProfile := profileFromDefNullable(requestData.Profile)

	// privacy
	requestPrivacy := privacyFromDefNullable(requestData.Privacy)

	//device
	requestDevice := requestData.Device

	noLoginParams, loginSession, mfaTypes, err := h.coreAPIs.Auth.Login(ip, string(requestDevice.Type), requestDevice.Os, requestDevice.DeviceId, string(requestData.AuthType),
		requestCreds, requestData.ApiKey, requestData.AppTypeIdentifier, requestData.OrgId, requestParams, &clientVersion, requestProfile, requestPrivacy, requestPreferences,
		requestData.AccountIdentifierId, false, l)
	if err != nil {
		loggingErr, ok := err.(*errors.Error)
		if ok && loggingErr.Status() != "" {
			return l.HTTPResponseError("Error logging in", err, http.StatusUnauthorized, true)
		}
		return l.HTTPResponseError("Error logging in", err, http.StatusInternalServerError, true)
	}

	///prepare response

	//noLoginParams
	if noLoginParams != nil {
		var message *string
		if messageVal, _ := noLoginParams["message"].(string); messageVal != "" {
			message = &messageVal
		}

		var paramsRes Def.SharedResLogin_Params
		paramsBytes, err := json.Marshal(noLoginParams)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.MessageDataType("no login response params"), nil, err, http.StatusInternalServerError, false)
		}

		err = json.Unmarshal(paramsBytes, &paramsRes)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("no login response params"), nil, err, http.StatusInternalServerError, false)
		}

		responseData := &Def.SharedResLogin{Message: message, Params: &paramsRes}
		respData, err := json.Marshal(responseData)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.MessageDataType("auth login response"), nil, err, http.StatusInternalServerError, false)
		}
		return l.HTTPResponseSuccessJSON(respData)
	}

	if loginSession.State != "" {
		paramsRes, err := utils.JSONConvert[Def.SharedResLoginMfa_Params](loginSession.Params)
		if err != nil {
			return l.HTTPResponseErrorAction("converting", logutils.MessageDataType("auth login response params"), nil, err, http.StatusInternalServerError, false)
		}

		mfaResp := mfaDataListToDef(mfaTypes)
		responseData := &Def.SharedResLoginMfa{AccountId: loginSession.Identifier, Enrolled: mfaResp, Params: paramsRes,
			SessionId: loginSession.ID, State: loginSession.State}
		respData, err := json.Marshal(responseData)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.MessageDataType("auth login response"), nil, err, http.StatusInternalServerError, false)
		}
		return l.HTTPResponseSuccessJSON(respData)
	}

	return authBuildLoginResponse(l, loginSession)
}

func (h ServicesApisHandler) loginMFA(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var mfaData Def.SharedReqLoginMfa
	err := json.NewDecoder(r.Body).Decode(&mfaData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("login mfa request"), nil, err, http.StatusBadRequest, true)
	}

	l.AddContext("account_id", mfaData.AccountId)
	message, loginSession, err := h.coreAPIs.Auth.LoginMFA(mfaData.ApiKey, mfaData.AccountId, mfaData.SessionId, mfaData.Identifier, string(mfaData.Type), mfaData.Code, mfaData.State, l)
	if message != nil {
		return l.HTTPResponseError(*message, err, http.StatusUnauthorized, false)
	}
	if err != nil {
		return l.HTTPResponseError("Error logging in", err, http.StatusInternalServerError, true)
	}

	return authBuildLoginResponse(l, loginSession)
}

func (h ServicesApisHandler) refresh(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	clientVersion := r.Header.Get("CLIENT_VERSION")

	var requestData Def.SharedReqRefresh
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("auth refresh request"), nil, err, http.StatusBadRequest, true)
	}

	loginSession, err := h.coreAPIs.Auth.Refresh(requestData.RefreshToken, requestData.ApiKey, &clientVersion, l)
	if err != nil {
		return l.HTTPResponseError("Error refreshing token", err, http.StatusInternalServerError, true)
	}
	if loginSession == nil {
		//if login session is null then unauthorized
		return l.HTTPResponseError(http.StatusText(http.StatusUnauthorized), nil, http.StatusUnauthorized, true)
	}

	accessToken := loginSession.AccessToken
	refreshToken := loginSession.CurrentRefreshToken()

	paramsRes, err := utils.JSONConvert[Def.SharedResRefresh_Params](loginSession.Params)
	if err != nil {
		return l.HTTPResponseErrorAction("converting", logutils.MessageDataType("auth refresh response params"), nil, err, http.StatusInternalServerError, false)
	}

	tokenType := Def.SharedResRokwireTokenTokenTypeBearer
	rokwireToken := Def.SharedResRokwireToken{AccessToken: &accessToken, RefreshToken: &refreshToken, TokenType: &tokenType}
	responseData := &Def.SharedResRefresh{Token: &rokwireToken, Params: paramsRes}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.MessageDataType("auth refresh response"), nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) loginURL(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqLoginUrl
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "auth login url request", nil, err, http.StatusBadRequest, true)
	}

	loginURL, params, err := h.coreAPIs.Auth.GetLoginURL(string(requestData.AuthType), requestData.AppTypeIdentifier, requestData.OrgId, requestData.RedirectUri, requestData.ApiKey, l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "login url", nil, err, http.StatusInternalServerError, true)
	}

	responseData := &Def.SharedResLoginUrl{LoginUrl: loginURL, Params: &params}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "auth login url response", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) accountExists(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqAccountCheck
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequest, nil, err, http.StatusBadRequest, true)
	}

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	//auth type
	var authType *string
	if requestData.AuthType != nil {
		authTypeStr := string(*requestData.AuthType)
		authType = &authTypeStr
	}

	accountExists, err := h.coreAPIs.Auth.AccountExists(requestIdentifier, requestData.ApiKey, requestData.AppTypeIdentifier, requestData.OrgId, authType, requestData.UserIdentifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, logutils.MessageDataType("account exists"), nil, err, http.StatusInternalServerError, false)
	}

	respData, err := json.Marshal(accountExists)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) canSignIn(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqAccountCheck
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequest, nil, err, http.StatusBadRequest, true)
	}

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	//auth type
	var authType *string
	if requestData.AuthType != nil {
		authTypeStr := string(*requestData.AuthType)
		authType = &authTypeStr
	}

	canSignIn, err := h.coreAPIs.Auth.CanSignIn(requestIdentifier, requestData.ApiKey, requestData.AppTypeIdentifier, requestData.OrgId, authType, requestData.UserIdentifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, logutils.MessageDataType("can sign in"), nil, err, http.StatusInternalServerError, false)
	}

	respData, err := json.Marshal(canSignIn)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) canLink(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqAccountCheck
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequest, nil, err, http.StatusBadRequest, true)
	}

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	//auth type
	var authType *string
	if requestData.AuthType != nil {
		authTypeStr := string(*requestData.AuthType)
		authType = &authTypeStr
	}

	canLink, err := h.coreAPIs.Auth.CanLink(requestIdentifier, requestData.ApiKey, requestData.AppTypeIdentifier, requestData.OrgId, authType, requestData.UserIdentifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, logutils.MessageDataType("can link"), nil, err, http.StatusInternalServerError, false)
	}

	respData, err := json.Marshal(canLink)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) signInOptions(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqAccountCheck
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequest, nil, err, http.StatusBadRequest, true)
	}

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	//auth type
	var authType *string
	if requestData.AuthType != nil {
		authTypeStr := string(*requestData.AuthType)
		authType = &authTypeStr
	}

	identifiers, authTypes, err := h.coreAPIs.Auth.SignInOptions(requestIdentifier, requestData.ApiKey, requestData.AppTypeIdentifier, requestData.OrgId, authType, requestData.UserIdentifier, l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, logutils.MessageDataType("sign-in options"), nil, err, http.StatusInternalServerError, false)
	}

	respIdentifiers := accountIdentifiersToDef(identifiers)
	respAuthTypes := accountAuthTypesToDef(authTypes)
	resp := Def.SharedResSignInOptions{Identifiers: respIdentifiers, AuthTypes: respAuthTypes}

	respData, err := json.Marshal(resp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponse, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) linkAccountAuthType(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqAccountAuthTypeLink
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("account auth type link request"), nil, err, http.StatusBadRequest, true)
	}

	//creds
	requestCreds, err := interfaceToJSON(requestData.Creds)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	//params
	requestParams, err := interfaceToJSON(requestData.Params)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "params", nil, err, http.StatusBadRequest, true)
	}

	message, account, err := h.coreAPIs.Auth.LinkAccountAuthType(claims.Subject, string(requestData.AuthType), requestData.AppTypeIdentifier, requestCreds, requestParams, l)
	if err != nil {
		return l.HTTPResponseError("Error linking account auth type", err, http.StatusInternalServerError, true)
	}

	identifiers := make([]Def.AccountIdentifier, 0)
	authTypes := make([]Def.AccountAuthType, 0)
	if account != nil {
		account.SortAccountAuthTypes("", claims.AuthType)
		identifiers = accountIdentifiersToDef(account.Identifiers)
		authTypes = accountAuthTypesToDefLegacy(account)
	}

	responseData := &Def.ServicesResAccountAuthTypeLink{Identifiers: &identifiers, AuthTypes: authTypes, Message: message}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "link account auth type response", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) unlinkAccountAuthType(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqAccountAuthTypeUnlink
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("account auth type unlink request"), nil, err, http.StatusBadRequest, true)
	}

	var authType *string
	if requestData.AuthType != nil {
		authTypeStr := string(*requestData.AuthType)
		authType = &authTypeStr
	}
	account, err := h.coreAPIs.Auth.UnlinkAccountAuthType(claims.Subject, requestData.Id, authType, requestData.Identifier, false, l)
	if err != nil {
		return l.HTTPResponseError("Error unlinking account auth type", err, http.StatusInternalServerError, true)
	}

	identifiers := make([]Def.AccountIdentifier, 0)
	authTypes := make([]Def.AccountAuthType, 0)
	if account != nil {
		account.SortAccountAuthTypes("", claims.AuthType)
		identifiers = accountIdentifiersToDef(account.Identifiers)
		authTypes = accountAuthTypesToDefLegacy(account)
	}

	responseData := &Def.ServicesResAccountAuthTypeLink{Identifiers: &identifiers, AuthTypes: authTypes}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "unlink account auth type response", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) linkAccountIdentifier(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqAccountIdentifierLink
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("account identifier link request"), nil, err, http.StatusBadRequest, true)
	}

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	message, account, err := h.coreAPIs.Auth.LinkAccountIdentifier(claims.Subject, requestIdentifier, false, l)
	if err != nil {
		return l.HTTPResponseError("Error linking account identifier", err, http.StatusInternalServerError, true)
	}

	identifiers := make([]Def.AccountIdentifier, 0)
	if account != nil {
		identifiers = accountIdentifiersToDef(account.Identifiers)
	}

	responseData := &Def.ServicesResAccountIdentifierLink{Identifiers: identifiers, Message: message}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "link account identifier response", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) unlinkAccountIdentifier(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqAccountIdentifierUnlink
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("account identifier unlink request"), nil, err, http.StatusBadRequest, true)
	}

	account, err := h.coreAPIs.Auth.UnlinkAccountIdentifier(claims.Subject, requestData.Id, false, l)
	if err != nil {
		return l.HTTPResponseError("Error unlinking account identifier", err, http.StatusInternalServerError, true)
	}

	identifiers := make([]Def.AccountIdentifier, 0)
	if account != nil {
		identifiers = accountIdentifiersToDef(account.Identifiers)
	}

	responseData := &Def.ServicesResAccountIdentifierLink{Identifiers: identifiers}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "unlink account identifier response", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) authorizeService(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqAuthorizeService
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "auth authorize service request", nil, err, http.StatusBadRequest, true)
	}

	var scopes []authorization.Scope
	if requestData.ApprovedScopes != nil && *requestData.ApprovedScopes != nil {
		scopes, err = authorization.ScopesFromStrings(*requestData.ApprovedScopes, false)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, "scopes", nil, err, http.StatusBadRequest, true)
		}
	}

	//TODO: Fill "claims" with claims from access token
	token, tokenScopes, reg, err := h.coreAPIs.Auth.AuthorizeService(tokenauth.Claims{}, requestData.ServiceId, scopes, l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "login url", nil, err, http.StatusInternalServerError, true)
	}

	scopesResp := authorization.ScopesToStrings(tokenScopes)
	regResp := serviceRegToDef(reg)
	tokenType := Def.ServicesResAuthorizeServiceTokenTypeBearer

	responseData := &Def.ServicesResAuthorizeService{AccessToken: &token, TokenType: &tokenType, ApprovedScopes: &scopesResp, ServiceReg: regResp}
	respData, err := json.Marshal(responseData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "auth login url response", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) getServiceRegistrations(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	serviceIDsParam := r.URL.Query().Get("ids")
	if serviceIDsParam == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("ids"), nil, http.StatusBadRequest, false)
	}
	serviceIDs := strings.Split(serviceIDsParam, ",")

	serviceRegs := h.coreAPIs.Auth.GetServiceRegistrations(serviceIDs)
	serviceRegResp := serviceRegListToDef(serviceRegs)

	data, err := json.Marshal(serviceRegResp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeServiceReg, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) deleteAccount(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	err := h.coreAPIs.Services.SerDeleteAccount(claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) getAccount(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	account, err := h.coreAPIs.Services.SerGetAccount(claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	var accountData *Def.Account
	if account != nil {
		account.SortAccountAuthTypes("", claims.AuthType)
		accountData = accountToDef(*account)
	}

	data, err := json.Marshal(accountData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccount, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) createAdminAccount(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	clientVersion := r.Header.Get("CLIENT_VERSION")

	var requestData Def.SharedReqCreateAccount
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("create account request"), nil, err, http.StatusBadRequest, true)
	}

	var permissions []string
	if requestData.Permissions != nil {
		permissions = *requestData.Permissions
	}
	var roleIDs []string
	if requestData.RoleIds != nil {
		roleIDs = *requestData.RoleIds
	}
	var groupIDs []string
	if requestData.GroupIds != nil {
		groupIDs = *requestData.GroupIds
	}
	var scopes []string
	if requestData.Scopes != nil {
		scopes = *requestData.Scopes
	}
	profile := profileFromDefNullable(requestData.Profile)
	privacy := privacyFromDefNullable(requestData.Privacy)

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	creatorPermissions := strings.Split(claims.Permissions, ",")
	account, params, err := h.coreAPIs.Auth.CreateAdminAccount(string(requestData.AuthType), claims.AppID, claims.OrgID,
		requestIdentifier, profile, privacy, permissions, roleIDs, groupIDs, scopes, creatorPermissions, &clientVersion, l)
	if err != nil || account == nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	respData := partialAccountToDef(*account, params)

	data, err := json.Marshal(respData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccount, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) updateAdminAccount(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqUpdateAccount
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("update account request"), nil, err, http.StatusBadRequest, true)
	}

	var permissions []string
	if requestData.Permissions != nil {
		permissions = *requestData.Permissions
	}
	var roleIDs []string
	if requestData.RoleIds != nil {
		roleIDs = *requestData.RoleIds
	}
	var groupIDs []string
	if requestData.GroupIds != nil {
		groupIDs = *requestData.GroupIds
	}
	var scopes []string
	if requestData.Scopes != nil {
		scopes = *requestData.Scopes
	}

	//identifier
	requestIdentifier, err := interfaceToJSON(requestData.Identifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	updaterPermissions := strings.Split(claims.Permissions, ",")
	account, params, err := h.coreAPIs.Auth.UpdateAdminAccount(string(requestData.AuthType), claims.AppID, claims.OrgID, requestIdentifier,
		permissions, roleIDs, groupIDs, scopes, updaterPermissions, l)
	if err != nil || account == nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	respData := partialAccountToDef(*account, params)

	data, err := json.Marshal(respData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccount, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) getMFATypes(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	mfaDataList, err := h.coreAPIs.Auth.GetMFATypes(claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeMFAType, nil, err, http.StatusInternalServerError, true)
	}

	mfaResp := mfaDataListToDef(mfaDataList)

	data, err := json.Marshal(mfaResp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeMFAType, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) addMFAType(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var mfaData Def.SharedReqMfa
	err := json.NewDecoder(r.Body).Decode(&mfaData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("add mfa request"), nil, err, http.StatusBadRequest, true)
	}

	mfa, err := h.coreAPIs.Auth.AddMFAType(claims.Subject, mfaData.Identifier, string(mfaData.Type))
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionInsert, model.TypeMFAType, nil, err, http.StatusInternalServerError, true)
	}

	mfaResp := mfaDataToDef(mfa)

	respData, err := json.Marshal(mfaResp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeMFAType, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(respData)
}

func (h ServicesApisHandler) removeMFAType(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var mfaData Def.SharedReqMfa
	err := json.NewDecoder(r.Body).Decode(&mfaData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("remove mfa request"), nil, err, http.StatusBadRequest, true)
	}

	err = h.coreAPIs.Auth.RemoveMFAType(claims.Subject, mfaData.Identifier, string(mfaData.Type))
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeMFAType, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) getProfile(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	profile, err := h.coreAPIs.Services.SerGetProfile(claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeProfile, nil, err, http.StatusInternalServerError, true)
	}

	profileResp := profileToDef(profile)

	// maintain backwards compatibility
	if len(profile.Accounts) == 1 {
		if emailIdentifier := profile.Accounts[0].GetAccountIdentifier("email", ""); emailIdentifier != nil {
			profileResp.Email = &emailIdentifier.Identifier
		}
		if phoneIdentifier := profile.Accounts[0].GetAccountIdentifier("phone", ""); phoneIdentifier != nil {
			profileResp.Phone = &phoneIdentifier.Identifier
		}
	}

	data, err := json.Marshal(profileResp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeProfile, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) updateProfile(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.Profile
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "profile update request", nil, err, http.StatusBadRequest, true)
	}

	profile := profileFromDef(&requestData)

	err = h.coreAPIs.Services.SerUpdateAccountProfile(claims.Subject, profile)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeProfile, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) updatePrivacy(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {

	var requestData Def.Privacy
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	privacy := privacyFromDef(&requestData)

	err = h.coreAPIs.Services.SerUpdateAccountPrivacy(claims.Subject, privacy)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypePrivacy, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) updateAccountPreferences(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var preferences map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&preferences)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "account preferences update request", nil, err, http.StatusBadRequest, true)
	}

	created, err := h.coreAPIs.Services.SerUpdateAccountPreferences(claims.Subject, claims.AppID, claims.OrgID, claims.Anonymous, preferences, l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeAccountPreferences, nil, err, http.StatusInternalServerError, true)
	}

	if created {
		return l.HTTPResponseSuccessMessage("Created new anonymous account with ID " + claims.Subject)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) getPreferences(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	preferences, err := h.coreAPIs.Services.SerGetPreferences(claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAccountPreferences, nil, err, http.StatusInternalServerError, true)
	}

	response := preferences

	data, err := json.Marshal(response)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccountPreferences, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) updateAccountSecrets(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var secrets map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&secrets)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "account secrets update request", nil, err, http.StatusBadRequest, true)
	}

	err = h.coreAPIs.Services.SerUpdateAccountSecrets(claims.Subject, secrets)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeAccountSecrets, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) getAccountSystemConfigs(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	configs, err := h.coreAPIs.Services.SerGetAccountSystemConfigs(claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAccountSystemConfigs, nil, err, http.StatusInternalServerError, true)
	}

	response := configs

	data, err := json.Marshal(response)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccountSystemConfigs, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) updateAccountUsername(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var username Def.Username
	err := json.NewDecoder(r.Body).Decode(&username)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, model.TypeAccountUsername, nil, err, http.StatusBadRequest, true)
	}

	err = h.coreAPIs.Services.SerUpdateAccountUsername(claims.Subject, claims.AppID, claims.OrgID, username.Username)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeAccountUsername, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) getAccounts(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var err error

	//limit and offset
	limit := 20
	limitArg := r.URL.Query().Get("limit")
	if limitArg != "" {
		limit, err = strconv.Atoi(limitArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("limit"), err, http.StatusBadRequest, false)
		}
	}
	offset := 0
	offsetArg := r.URL.Query().Get("offset")
	if offsetArg != "" {
		offset, err = strconv.Atoi(offsetArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("offset"), err, http.StatusBadRequest, false)
		}
	}

	//account ID
	var accountID *string
	accountIDParam := r.URL.Query().Get("account-id")
	if len(accountIDParam) > 0 {
		accountID = &accountIDParam
	}
	//first name
	var firstName *string
	firstNameParam := r.URL.Query().Get("firstname")
	if len(firstNameParam) > 0 {
		firstName = &firstNameParam
	}
	//last name
	var lastName *string
	lastNameParam := r.URL.Query().Get("lastname")
	if len(lastNameParam) > 0 {
		lastName = &lastNameParam
	}
	//auth type
	var authType *string
	authTypeParam := r.URL.Query().Get("auth-type")
	if len(authTypeParam) > 0 {
		authType = &authTypeParam
	}
	//auth type identifier
	var authTypeIdentifier *string
	authTypeIdentifierParam := r.URL.Query().Get("auth-type-identifier")
	if len(authTypeIdentifierParam) > 0 {
		authTypeIdentifier = &authTypeIdentifierParam
	}

	//has permissions
	var hasPermissions *bool
	hasPermissionsArg := r.URL.Query().Get("has-permissions")
	if hasPermissionsArg != "" {
		hasPermissionsVal, err := strconv.ParseBool(hasPermissionsArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("has-permissions"), err, http.StatusBadRequest, false)
		}
		hasPermissions = &hasPermissionsVal
	}
	//anonymous
	var anonymous *bool
	anonymousArg := r.URL.Query().Get("anonymous")
	if anonymousArg != "" {
		anonymousVal, err := strconv.ParseBool(anonymousArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("anonymous"), err, http.StatusBadRequest, false)
		}
		anonymous = &anonymousVal
	}
	//permissions
	var permissions []string
	permissionsArg := r.URL.Query().Get("permissions")
	if permissionsArg != "" {
		permissions = strings.Split(permissionsArg, ",")
	}
	//roleIDs
	var roleIDs []string
	rolesArg := r.URL.Query().Get("role-ids")
	if rolesArg != "" {
		roleIDs = strings.Split(rolesArg, ",")
	}
	//groupIDs
	var groupIDs []string
	groupsArg := r.URL.Query().Get("group-ids")
	if groupsArg != "" {
		groupIDs = strings.Split(groupsArg, ",")
	}

	accounts, err := h.coreAPIs.Services.SerGetAccounts(limit, offset, claims.AppID, claims.OrgID, accountID, firstName, lastName, authType, authTypeIdentifier, anonymous, hasPermissions, permissions, roleIDs, groupIDs)
	if err != nil {
		return l.HTTPResponseErrorAction("error finding accounts", model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	response := partialAccountsToDef(accounts)

	data, err := json.Marshal(response)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccount, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) getPublicAccounts(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var err error

	//limit and offset
	limit := 20
	limitArg := r.URL.Query().Get("limit")
	if limitArg != "" {
		limit, err = strconv.Atoi(limitArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("limit"), err, http.StatusBadRequest, false)
		}
	}
	offset := 0
	offsetArg := r.URL.Query().Get("offset")
	if offsetArg != "" {
		offset, err = strconv.Atoi(offsetArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("offset"), err, http.StatusBadRequest, false)
		}
	}

	//search
	var search *string
	searchParam := r.URL.Query().Get("search")
	if len(searchParam) > 0 {
		search = &searchParam
	}

	//username
	var username *string
	usernameParam := r.URL.Query().Get("username")
	if len(usernameParam) > 0 {
		username = &usernameParam
	}

	//first name
	var firstName *string
	firstNameParam := r.URL.Query().Get("firstname")
	if len(firstNameParam) > 0 {
		firstName = &firstNameParam
	}
	//last name
	var lastName *string
	lastNameParam := r.URL.Query().Get("lastname")
	if len(lastNameParam) > 0 {
		lastName = &lastNameParam
	}

	//following id
	var followingID *string
	followingIDParam := r.URL.Query().Get("following-id")
	if len(followingIDParam) > 0 {
		followingID = &followingIDParam
	}

	//following id
	var followerID *string
	followerIDParam := r.URL.Query().Get("follower-id")
	if len(followerIDParam) > 0 {
		followerID = &followerIDParam
	}

	accounts, err := h.coreAPIs.Services.SerGetPublicAccounts(claims.AppID, claims.OrgID, limit, offset, search,
		firstName, lastName, username, followingID, followerID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction("error finding accounts", model.TypeAccount, nil, err, http.StatusInternalServerError, true)
	}

	if accounts == nil {
		accounts = []model.PublicAccount{}
	}

	data, err := json.Marshal(accounts)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAccount, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(data)
}

func (h ServicesApisHandler) addFollow(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var follow Def.Follow
	err := json.NewDecoder(r.Body).Decode(&follow)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, model.TypeFollow, nil, err, http.StatusBadRequest, true)
	}

	// Don't allow people to follow themselves
	if follow.FollowingId == claims.Subject {
		return l.HTTPResponseErrorAction(logutils.ActionInsert, model.TypeFollow, nil, err, http.StatusBadRequest, true)
	}

	// Check to make sure account exists
	account, err := h.coreAPIs.Services.SerGetAccount(follow.FollowingId)
	if err != nil || account == nil {
		return l.HTTPResponseErrorAction(logutils.ActionInsert, model.TypeFollow, nil, err, http.StatusBadRequest, true)
	}

	// Check to make sure follower account is public
	followerAccount, err := h.coreAPIs.Services.SerGetAccount(claims.Subject)
	if err != nil || followerAccount == nil || !followerAccount.Privacy.Public {
		return l.HTTPResponseErrorAction(logutils.ActionInsert, model.TypeFollow, nil, err, http.StatusBadRequest, true)
	}

	// Check to make sure account is public
	if !account.Privacy.Public {
		return l.HTTPResponseErrorAction(logutils.ActionInsert, model.TypeFollow, nil, err, http.StatusForbidden, true)
	}

	err = h.coreAPIs.Services.SerAddFollow(model.Follow{
		AppID:       claims.AppID,
		OrgID:       claims.OrgID,
		FollowerID:  claims.Subject,
		FollowingID: follow.FollowingId,
	})
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeFollow, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ServicesApisHandler) deleteFollow(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	followingID := params["id"]
	if len(followingID) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.coreAPIs.Services.SerDeleteFollow(claims.AppID, claims.OrgID, followingID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeFollow, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// getCommonTest TODO get test
func (h ServicesApisHandler) getTest(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	res := h.coreAPIs.Services.SerGetCommonTest(l)

	return l.HTTPResponseSuccessMessage(res)
}

func (h ServicesApisHandler) getApplicationConfigs(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqAppConfigs
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("application config request"), nil, err, http.StatusBadRequest, true)
	}

	version := model.VersionNumbersFromString(requestData.Version)
	if version == nil {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, model.TypeVersionNumbers, nil, nil, http.StatusBadRequest, false)
	}

	appConfig, err := h.coreAPIs.Services.SerGetAppConfig(requestData.AppTypeIdentifier, nil, *version, &requestData.ApiKey)
	if err != nil || appConfig == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeApplicationConfig, nil, err, http.StatusInternalServerError, true)
	}

	appConfigResp := appConfigToDef(*appConfig)

	response, err := json.Marshal(appConfigResp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeApplicationConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(response)
}

func (h ServicesApisHandler) getAppAssetFile(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	orgID := params["org_id"]
	if len(orgID) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("org_id"), nil, http.StatusBadRequest, false)
	}
	appID := params["app_id"]
	if len(appID) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("app_id"), nil, http.StatusBadRequest, false)
	}
	name := params["name"]
	if len(name) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("name"), nil, http.StatusBadRequest, false)
	}

	asset, err := h.coreAPIs.Services.SerGetAppAssetFile(orgID, appID, name)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppAsset, nil, err, http.StatusInternalServerError, true)
	}

	if asset == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAppAsset, nil, err, http.StatusNotFound, true)
	}

	response, err := json.Marshal(asset.Data)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeAppAsset, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(response)
}

func (h ServicesApisHandler) getApplicationOrgConfigs(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.SharedReqAppConfigsOrg
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("application org config request"), nil, err, http.StatusBadRequest, true)
	}

	version := model.VersionNumbersFromString(requestData.Version)
	if version == nil {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, model.TypeVersionNumbers, nil, nil, http.StatusBadRequest, false)
	}

	appConfig, err := h.coreAPIs.Services.SerGetAppConfig(requestData.AppTypeIdentifier, &claims.OrgID, *version, nil)
	if err != nil || appConfig == nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeApplicationConfig, nil, err, http.StatusInternalServerError, true)
	}

	appConfigResp := appConfigToDef(*appConfig)

	response, err := json.Marshal(appConfigResp)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeApplicationConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(response)
}

// Handler for reset credential endpoint from client application
func (h ServicesApisHandler) updateCredential(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqCredentialUpdate
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("auth reset credential client request"), nil, err, http.StatusBadRequest, true)
	}

	//params
	requestParams, err := interfaceToJSON(requestData.Params)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "params", nil, err, http.StatusBadRequest, true)
	}

	if err := h.coreAPIs.Auth.UpdateCredential(claims.Subject, requestData.AccountAuthTypeId, requestParams, l); err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "credential", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessMessage("Reset credential from client successfully")
}

// Handler for reset credential endpoint from reset link
func (h ServicesApisHandler) forgotCredentialComplete(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqCredentialForgotComplete
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("auth reset credential link request"), nil, err, http.StatusBadRequest, true)
	}

	//params
	requestParams, err := interfaceToJSON(requestData.Params)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "params", nil, err, http.StatusBadRequest, true)
	}

	if err := h.coreAPIs.Auth.ResetForgotCredential(requestData.CredentialId, requestData.ResetCode, requestParams, l); err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, "credential", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessMessage("Reset credential from link successfully")
}

// Handler for forgot credential endpoint
func (h ServicesApisHandler) forgotCredentialInitiate(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqCredentialForgotInitiate
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("auth reset credential request"), nil, err, http.StatusBadRequest, true)
	}

	var requestIdentifier interface{}
	if identifier, err := requestData.Identifier.AsSharedReqIdentifierString(); err == nil {
		requestIdentifier = map[string]string{string(requestData.AuthType): identifier}
	} else if identifier, err := requestData.Identifier.AsSharedReqIdentifiers(); err == nil {
		requestIdentifier = identifier
	} else {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.MessageDataType("auth reset credential identifier"), nil, err, http.StatusBadRequest, true)
	}

	//identifier
	identifierJSON, err := interfaceToJSON(requestIdentifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	if err := h.coreAPIs.Auth.ForgotCredential(string(requestData.AuthType), identifierJSON, requestData.AppTypeIdentifier, requestData.OrgId, requestData.ApiKey, l); err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "forgot credential link", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessMessage("Sent forgot credential link successfully")
}

// Handler for verify endpoint
func (h ServicesApisHandler) verifyIdentifier(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	id := r.URL.Query().Get("id")
	if id == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("code"), nil, http.StatusBadRequest, false)
	}

	accountIdentifier, err := h.coreAPIs.Auth.VerifyIdentifier(id, code, l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionValidate, model.TypeAccountIdentifier, nil, err, http.StatusInternalServerError, false)
	}

	identifierStr := "Account identifier"
	if accountIdentifier != nil && accountIdentifier.Code != "" {
		identifierStr = cases.Title(language.English).String(accountIdentifier.Code)
	}
	return l.HTTPResponseSuccessMessage(fmt.Sprintf("%s verified successfully!", identifierStr))
}

// Handler for resending verify code
func (h ServicesApisHandler) sendVerifyIdentifier(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.ServicesReqIdentifierSendVerify
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("auth resend verification request"), nil, err, http.StatusBadRequest, true)
	}

	var requestIdentifier interface{}
	if identifier, err := requestData.Identifier.AsSharedReqIdentifierString(); err == nil && requestData.AuthType != nil {
		authType := string(*requestData.AuthType)
		requestIdentifier = map[string]string{authType: identifier}
	} else if identifier, err := requestData.Identifier.AsSharedReqIdentifiers(); err == nil {
		requestIdentifier = identifier
	} else {
		return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.MessageDataType("auth resend verification identifier"), nil, err, http.StatusBadRequest, true)
	}

	//identifier
	identifierJSON, err := interfaceToJSON(requestIdentifier)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeCreds, nil, err, http.StatusBadRequest, true)
	}

	if err := h.coreAPIs.Auth.SendVerifyIdentifier(requestData.AppTypeIdentifier, requestData.OrgId, requestData.ApiKey, identifierJSON, l); err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionSend, "code", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessMessage("Verification code sent")
}

func (h ServicesApisHandler) verifyMFA(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var mfaData Def.SharedReqMfa
	err := json.NewDecoder(r.Body).Decode(&mfaData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("verify mfa request"), nil, err, http.StatusBadRequest, true)
	}

	if mfaData.Code == nil || *mfaData.Code == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, "mfa code", nil, nil, http.StatusBadRequest, true)
	}

	message, recoveryCodes, err := h.coreAPIs.Auth.VerifyMFA(claims.Subject, mfaData.Identifier, string(mfaData.Type), *mfaData.Code)
	if message != nil {
		return l.HTTPResponseError(*message, nil, http.StatusBadRequest, true)
	}
	if err != nil {
		return l.HTTPResponseError("Error verifying MFA", err, http.StatusInternalServerError, true)
	}

	if recoveryCodes == nil {
		recoveryCodes = []string{}
	}

	response, err := json.Marshal(recoveryCodes)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "mfa recovery codes", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(response)
}

func (h ServicesApisHandler) logout(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestDataData Def.PostServicesAuthLogoutJSONBody
	err := json.NewDecoder(r.Body).Decode(&requestDataData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.MessageDataType("verify logout request"), nil, err, http.StatusBadRequest, true)
	}

	err = h.coreAPIs.Auth.Logout(claims.AppID, claims.OrgID, claims.Subject, claims.SessionID, requestDataData.AllSessions, l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeLoginSession, nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// NewServicesApisHandler creates new rest services Handler instance
func NewServicesApisHandler(coreAPIs *core.APIs) ServicesApisHandler {
	return ServicesApisHandler{coreAPIs: coreAPIs}
}

// HTMLResponseTemplate represents html response template
type HTMLResponseTemplate struct {
	Message string
}
