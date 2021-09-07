package auth

import (
	"core-building-block/core/model"
	"strings"

	"github.com/rokmetro/auth-library/authorization"
	"github.com/rokmetro/auth-library/authutils"
	"github.com/rokmetro/logging-library/errors"
	"github.com/rokmetro/logging-library/logutils"

	"github.com/rokmetro/logging-library/logs"
)

//Start starts the auth service
func (a *Auth) Start() {
	storageListener := StorageListener{auth: a}
	a.storage.RegisterStorageListener(&storageListener)

	go a.setupDeleteRefreshTimer()
}

//GetHost returns the host/issuer of the auth service
func (a *Auth) GetHost() string {
	return a.host
}

//Login logs a user in a specific application using the specified credentials and authentication method.
//The authentication method must be one of the supported for the application.
//	Input:
//		authType (string): Name of the authentication method for provided creds (eg. "email", "username", "illinois_oidc")
//		creds (string): Credentials/JSON encoded credential structure defined for the specified auth type
//		appID (string): ID of the app/client that the user is logging in from
//		params (string): JSON encoded params defined by specified auth type
//		l (*logs.Log): Log object pointer for request
//	Returns:
//		Access token (string): Signed ROKWIRE access token to be used to authorize future requests
//		Refresh Token (string): Refresh token that can be sent to refresh the access token once it expires
//		User (User): User object for authenticated user
//		Params (interface{}): authType-specific set of parameters passed back to client
func (a *Auth) Login(authType string, creds string, appID string, params string, l *logs.Log) (string, string, *model.Account, interface{}, error) {
	//validate if the provided auth type is supported by the provided application

	return "", "", nil, nil, nil
	/*	authTypeEntity, appTypeEntity, err := a.validateAuthType(authType, appID)
		if err != nil {
			return "", "", nil, nil, errors.WrapErrorAction(logutils.ActionValidate, typeAuthType, nil, err)
		}

		var user *model.User
		var userAuthType *model.UserAuthType

		//get the auth type implementation for the auth type
		if authTypeEntity.IsExternal {
			user, userAuthType, err = a.applyExternalAuthType(*authTypeEntity, creds, *appTypeEntity, params, l)
			if err != nil {
				return "", "", nil, nil, errors.WrapErrorAction("apply external auth type", "user", nil, err)
			}

			//TODO groups mapping

		} else {
			user, userAuthType, err = a.applyAuthType(*authTypeEntity, creds, *appTypeEntity, params, l)
			if err != nil {
				return "", "", nil, nil, errors.WrapErrorAction("apply auth type", "user", nil, err)
			}

			//the credentials are valid
		}

		//now we are ready to apply login for the user
		//TODO - get authTypeParam from the auth type.. pass nil for now
		authTypeParam := "TODO"

		accessToken, refreshToken, err := a.applyLogin(*user, *userAuthType, authTypeParam, l)
		if err != nil {
			return "", "", nil, nil, errors.WrapErrorAction("error apply login auth type", "user", nil, err)
		}

		return *accessToken, *refreshToken, user, authTypeParam, nil */
}

//Refresh refreshes an access token using a refresh token
//	Input:
//		refreshToken (string): Refresh token
//		l (*logs.Log): Log object pointer for request
//	Returns:
//		Access token (string): Signed ROKWIRE access token to be used to authorize future requests
//		Refresh token (string): Refresh token that can be sent to refresh the access token once it expires
//		Params (interface{}): authType-specific set of parameters passed back to client
func (a *Auth) Refresh(refreshToken string, l *logs.Log) (string, string, interface{}, error) {
	//TODO
	return "", "", nil, nil
	/*
		refresh, err := a.storage.FindRefreshToken(refreshToken)
		if err != nil {
			return "", "", nil, errors.WrapErrorAction("refreshing", logutils.TypeToken, nil, err)
		}

		credentials, err := a.storage.FindCredentialsByID(refresh.CredsID)
		if err != nil {
			return "", "", nil, errors.WrapErrorAction("refreshing", logutils.TypeToken, nil, err)
		}
		if credentials == nil {
			return "", "", nil, errors.ErrorData(logutils.StatusMissing, model.TypeAuthCred, nil)
		}

		validate := validator.New()
		err = validate.Struct(refresh)
		if err != nil {
			return "", "", nil, errors.WrapErrorAction(logutils.ActionValidate, typeAuthRefreshParams, nil, err)
		}

		if !refresh.Expires.After(time.Now().UTC()) {
			err = a.storage.DeleteRefreshToken(refresh.CurrentToken)
			if err != nil {
				return "", "", nil, errors.WrapErrorAction(logutils.ActionValidate, "refresh expiration", nil, err)
			}
			return "", "", nil, errors.ErrorAction(logutils.ActionValidate, "refresh expiration", nil)
		}

		if refreshToken == refresh.PreviousToken {
			err = a.storage.DeleteRefreshToken(refresh.CurrentToken)
			if err != nil {
				return "", "", nil, errors.WrapErrorAction(logutils.ActionValidate, "refresh reuse", nil, err)
			}
			return "", "", nil, errors.ErrorAction(logutils.ActionValidate, "refresh reuse", nil)
		}
		if refreshToken != refresh.CurrentToken {
			return "", "", nil, errors.ErrorAction(logutils.ActionValidate, model.TypeRefreshToken, nil)
		}

		auth, err := a.getAuthType(credentials.AuthType)
		if err != nil {
			return "", "", nil, errors.WrapErrorAction(logutils.ActionLoadCache, typeAuthType, nil, err)
		}

		userAuth, err := auth.refresh(refresh.Params, refresh.OrgID, refresh.AppID, l)
		if err != nil {
			return "", "", nil, errors.WrapErrorAction("refreshing", logutils.TypeToken, nil, err)
		}

		if userAuth == nil {
			return "", "", nil, errors.WrapErrorData(logutils.StatusInvalid, model.TypeUserAuth, nil, err)
		}

		user, err := a.storage.FindUserByAccountID(credentials.AccountID)
		if err != nil {
			return "", "", nil, err
		}

		user, update, _ := a.needsUserUpdate(userAuth, user)
		if update {
			_, err = a.updateAccount(user, "", nil)
			if err != nil {
				return "", "", nil, err
			}
		}

		claims := a.getStandardClaims(user.ID, userAuth.UserID, user.Account.Email, user.Account.Phone, "rokwire", refresh.OrgID, refresh.AppID, userAuth.Exp)
		token, err := a.buildAccessToken(claims, "", authorization.ScopeGlobal)
		if err != nil {
			return "", "", nil, errors.WrapErrorAction(logutils.ActionCreate, logutils.TypeToken, nil, err)
		}

		newRefreshToken := ""
		if userAuth.RefreshParams != nil {
			var expireTime *time.Time
			newRefreshToken, expireTime, err = a.buildRefreshToken()
			if err != nil {
				return "", "", nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeRefreshToken, nil, err)
			}

			now := time.Now().UTC()
			updatedRefresh := model.AuthRefresh{CurrentToken: newRefreshToken, PreviousToken: refreshToken, Expires: expireTime,
				Params: userAuth.RefreshParams, DateUpdated: &now}

			err = a.storage.UpdateRefreshToken(refreshToken, &updatedRefresh)
			if err != nil {
				return "", "", nil, err
			}
		}

		return token, newRefreshToken, userAuth.ResponseParams, nil
	*/
}

//GetLoginURL returns a pre-formatted login url for SSO providers
//	Input:
//		authenticationType (string): Name of the authentication method for provided creds (eg. "email", "username", "illinois_oidc")
//		appID (string): ID of the app/client that the user is logging in from
//		orgID (string): ID of the organization that the user is logging in from
//		redirectURI (string): Registered redirect URI where client will receive response
//		l (*loglib.Log): Log object pointer for request
//	Returns:
//		Login URL (string): SSO provider login URL to be launched in a browser
//		Params (map[string]interface{}): Params to be sent in subsequent request (if necessary)
func (a *Auth) GetLoginURL(authenticationType string, appID string, orgID string, redirectURI string, l *logs.Log) (string, map[string]interface{}, error) {
	//validate if the provided auth type is supported by the provided application and organization
	authType, appType, appOrg, err := a.validateAuthType(authenticationType, appID, orgID)
	if err != nil {
		return "", nil, errors.WrapErrorAction(logutils.ActionValidate, typeAuthType, nil, err)
	}

	//get the auth type implementation for the auth type
	authImpl, err := a.getExternalAuthTypeImpl(*authType)
	if err != nil {
		return "", nil, errors.WrapErrorAction(logutils.ActionLoadCache, typeAuthType, nil, err)
	}

	//get login URL
	loginURL, params, err := authImpl.getLoginURL(*authType, *appType, *appOrg, redirectURI, l)
	if err != nil {
		return "", nil, errors.WrapErrorAction(logutils.ActionGet, "login url", nil, err)
	}

	return loginURL, params, nil
}

//AuthorizeService returns a scoped token for the specified service and the service registration record if authorized or
//	the service registration record if not. Passing "approvedScopes" will update the service authorization for this user and
//	return a scoped access token which reflects this change.
//	Input:
//		claims (tokenClaims): Claims from un-scoped user access token
//		serviceID (string): ID of the service to be authorized
//		approvedScopes ([]string): list of scope strings to be approved
//		l (*logs.Log): Log object pointer for request
//	Returns:
//		Access token (string): Signed scoped access token to be used to authorize requests to the specified service
//		Approved Scopes ([]authorization.Scope): The approved scopes included in the provided token
//		Service reg (*model.ServiceReg): The service registration record for the requested service
func (a *Auth) AuthorizeService(claims TokenClaims, serviceID string, approvedScopes []authorization.Scope, l *logs.Log) (string, []authorization.Scope, *model.ServiceReg, error) {
	var authorization model.ServiceAuthorization
	if approvedScopes != nil {
		//If approved scopes are being updated, save update and return token with updated scopes
		authorization = model.ServiceAuthorization{UserID: claims.Subject, ServiceID: serviceID, Scopes: approvedScopes}
		err := a.storage.SaveServiceAuthorization(&authorization)
		if err != nil {
			return "", nil, nil, errors.WrapErrorAction(logutils.ActionSave, model.TypeServiceAuthorization, nil, err)
		}
	} else {
		serviceAuth, err := a.storage.FindServiceAuthorization(claims.Subject, serviceID)
		if err != nil {
			return "", nil, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeServiceAuthorization, nil, err)
		}

		if serviceAuth != nil {
			//If service authorization exists, generate token with saved scopes
			authorization = *serviceAuth
		} else {
			//If no service authorization exists, return the service registration record
			reg, err := a.storage.FindServiceReg(serviceID)
			if err != nil {
				return "", nil, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeServiceReg, nil, err)
			}
			return "", nil, reg, nil
		}
	}

	token, err := a.GetScopedAccessToken(claims, serviceID, authorization.Scopes)
	if err != nil {
		return "", nil, nil, errors.WrapErrorAction("build", logutils.TypeToken, nil, err)
	}

	return token, authorization.Scopes, nil, nil
}

//GetScopedAccessToken returns a scoped access token with the requested scopes
func (a *Auth) GetScopedAccessToken(claims TokenClaims, serviceID string, scopes []authorization.Scope) (string, error) {
	scopeStrings := []string{}
	services := []string{serviceID}
	for _, scope := range scopes {
		scopeStrings = append(scopeStrings, scope.String())
		if !authutils.ContainsString(services, scope.ServiceID) {
			services = append(services, scope.ServiceID)
		}
	}

	aud := strings.Join(services, ",")
	scope := strings.Join(scopeStrings, " ")

	scopedClaims := a.getStandardClaims(claims.Subject, "", "", "", aud, claims.OrgID, claims.AppID, nil)
	return a.buildAccessToken(scopedClaims, "", scope)
}

//GetServiceRegistrations retrieves all service registrations
func (a *Auth) GetServiceRegistrations(serviceIDs []string) ([]model.ServiceReg, error) {
	return a.storage.FindServiceRegs(serviceIDs)
}

//RegisterService creates a new service registration
func (a *Auth) RegisterService(reg *model.ServiceReg) error {
	if reg != nil && !reg.FirstParty && strings.Contains(strings.ToUpper(reg.Name), rokwireKeyword) {
		return errors.Newf("the name of a third-party service may not contain \"%s\"", rokwireKeyword)
	}
	return a.storage.InsertServiceReg(reg)
}

//UpdateServiceRegistration updates an existing service registration
func (a *Auth) UpdateServiceRegistration(reg *model.ServiceReg) error {
	if reg != nil {
		if reg.Registration.ServiceID == authServiceID || reg.Registration.ServiceID == a.serviceID {
			return errors.Newf("modifying service registration not allowed for service id %v", reg.Registration.ServiceID)
		}
		if !reg.FirstParty && strings.Contains(strings.ToUpper(reg.Name), rokwireKeyword) {
			return errors.Newf("the name of a third-party service may not contain \"%s\"", rokwireKeyword)
		}
	}
	return a.storage.UpdateServiceReg(reg)
}

//DeregisterService deletes an existing service registration
func (a *Auth) DeregisterService(serviceID string) error {
	if serviceID == authServiceID || serviceID == a.serviceID {
		return errors.Newf("deregistering service not allowed for service id %v", serviceID)
	}
	return a.storage.DeleteServiceReg(serviceID)
}

//GetAuthKeySet generates a JSON Web Key Set for auth service registration
func (a *Auth) GetAuthKeySet() (*model.JSONWebKeySet, error) {
	authReg, err := a.AuthService.GetServiceReg("auth")
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionLoadCache, model.TypeServiceReg, logutils.StringArgs("auth"), err)
	}

	if authReg == nil || authReg.PubKey == nil || authReg.PubKey.Key == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypePubKey, nil)
	}

	jwk, err := model.JSONWebKeyFromPubKey(authReg.PubKey)
	if err != nil || jwk == nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeJSONWebKey, nil, err)
	}

	return &model.JSONWebKeySet{Keys: []model.JSONWebKey{*jwk}}, nil
}
