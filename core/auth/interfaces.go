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

package auth

import (
	"core-building-block/core/model"
	"core-building-block/driven/storage"
	"net/url"
	"time"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/rokwire/core-auth-library-go/v3/authorization"
	"github.com/rokwire/core-auth-library-go/v3/sigauth"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// identifierType is the interface for auth identifiers that are not external to the system
type identifierType interface {
	//getType returns the identifier code
	// Returns:
	//	identifierCode (string): identifier code
	getCode() string

	//getIdentifier returns the user identifier
	// Returns:
	//	userIdentifier (string): User identifier
	getIdentifier() string

	//withIdentifier parses the credentials and copies the calling identifierType while caching the identifier
	// Returns:
	//	identifierImpl (identifierType): Copy of calling identifierType with cached identifier
	withIdentifier(creds string) (identifierType, error)

	//buildIdentifier creates a new account identifier
	// Returns:
	//	message (string): response message
	//	accountIdentifier (*model.AccountIdentifier): the new account identifier
	buildIdentifier(accountID *string, appName string) (string, *model.AccountIdentifier, error)

	// masks the cached identifier
	maskIdentifier() (string, error)

	// gives whether the identifier must be verified before sign-in is allowed
	requireVerificationForSignIn() bool

	// checks whether the given account identifier is verified, restarts verification if necessary and possible
	checkVerified(accountIdentifier *model.AccountIdentifier, appName string) error

	//allowMultiple says whether an account may have multiple identifiers of this type
	// Returns:
	//	allowed (bool): whether mulitple identifier types are allowed
	allowMultiple() bool
}

type authCommunicationChannel interface {
	//verifies identifier (e.g., checks the verification code generated on email signup for email auth type)
	verifyIdentifier(accountIdentifier *model.AccountIdentifier, verification string) error

	//sends the verification code to the identifier
	// Returns:
	//	sentCode (bool): whether the verification code was sent successfully
	sendVerifyIdentifier(accountIdentifier *model.AccountIdentifier, appName string) (bool, error)

	//restarts the identifier verification
	restartIdentifierVerification(accountIdentifier *model.AccountIdentifier, appName string) error

	//sendCode sends a verification code using the channel
	// Returns:
	//	message (string): response message
	sendCode(appName string, code string, codeType string, itemID string) (string, error)

	//requiresCodeGeneration says whether a code needs to be generated by this service to send on the channel
	// Returns:
	//	required (bool): whether codes need to be generated by the service
	requiresCodeGeneration() bool
}

// authType is the interface for authentication for auth types which are not external for the system(the users do not come from external system)
type authType interface {
	//signUp applies sign up operation
	// Returns:
	//	message (string): Success message if verification is required. If verification is not required, return ""
	//	accountIdentifier (*model.AccountIdentifier): new account identifier
	//	credential (*model.Credential): new credential
	signUp(identifierImpl identifierType, accountID *string, appOrg model.ApplicationOrganization, creds string, params string) (string, *model.AccountIdentifier, *model.Credential, error)

	//signUpAdmin signs up a new admin user
	// Returns:
	//	credentialParams (map): newly generated credential parameters
	//	accountIdentifier (*model.AccountIdentifier): new account identifier
	//	credential (*model.Credential): new credential
	signUpAdmin(identifierImpl identifierType, appOrg model.ApplicationOrganization, creds string) (map[string]interface{}, *model.AccountIdentifier, *model.Credential, error)

	//apply forgot credential for the auth type (generates a reset password link with code and expiry and sends it to given identifier for email auth type)
	forgotCredential(identifierImpl identifierType, credential *model.Credential, appOrg model.ApplicationOrganization) (map[string]interface{}, error)

	//updates the value of the credential object with new value
	// Returns:
	//	authTypeCreds (map[string]interface{}): Updated Credential.Value
	resetCredential(credential *model.Credential, resetCode *string, params string) (map[string]interface{}, error)

	//checkCredential checks if the incoming credentials are valid for the stored credentials
	// Returns:
	//	message (string): information required to complete login, if applicable
	//	credentialID (string): the ID of the credential used to validate the login
	checkCredentials(identifierImpl identifierType, accountID *string, aats []model.AccountAuthType, creds string, params string, appOrg model.ApplicationOrganization) (string, string, error)

	//withParams parses the params and copies the calling authType while caching the params
	// Returns:
	//	authImpl (authType): Copy of calling authType with cached params
	withParams(params map[string]interface{}) (authType, error)

	// gives whether the identifier used with this auth type must be verified before sign-in is allowed
	requireIdentifierVerificationForSignIn() bool

	//allowMultiple says whether an account may have multiple auth types of this type
	// Returns:
	//	allowed (bool): whether mulitple auth types are allowed
	allowMultiple() bool
}

// externalAuthType is the interface for authentication for auth types which are external for the system(the users comes from external system).
// these are the different identity providers - illinois_oidc etc
type externalAuthType interface {
	//getLoginUrl retrieves and pre-formats a login url and params for the SSO provider
	getLoginURL(authType model.AuthType, appType model.ApplicationType, redirectURI string, l *logs.Log) (string, map[string]interface{}, error)
	//externalLogin logins in the external system and provides the authenticated user
	externalLogin(authType model.AuthType, appType model.ApplicationType, appOrg model.ApplicationOrganization, creds string, params string, l *logs.Log) (*model.ExternalSystemUser, map[string]interface{}, string, error)
	//refresh refreshes tokens
	refresh(params map[string]interface{}, authType model.AuthType, appType model.ApplicationType, appOrg model.ApplicationOrganization, l *logs.Log) (*model.ExternalSystemUser, map[string]interface{}, string, error)
}

// anonymousAuthType is the interface for authentication for auth types which are anonymous
type anonymousAuthType interface {
	//checkCredentials checks the credentials for the provided app and organization
	//	Returns anonymous profile identifier
	checkCredentials(creds string) (string, map[string]interface{}, error)
}

// serviceAuthType is the interface for authentication for non-human clients
type serviceAuthType interface {
	checkCredentials(r *sigauth.Request, creds interface{}, params map[string]interface{}) ([]model.ServiceAccount, error)
	addCredentials(creds *model.ServiceAccountCredential) (map[string]interface{}, error)
}

// mfaType is the interface for multi-factor authentication
type mfaType interface {
	//verify verifies the code based on stored mfa params
	verify(context storage.TransactionContext, mfa *model.MFAType, accountID string, code string) (*string, error)
	//enroll creates a mfa type to be added to an account
	enroll(identifier string) (*model.MFAType, error)
	//sendCode generates a mfa code and expiration time and sends the code to the user
	sendCode(identifier string) (string, *time.Time, error)
}

// APIs is the interface which defines the APIs provided by the auth package
type APIs interface {
	//Start starts the auth service
	Start()

	//GetHost returns the host/issuer of the auth service
	GetHost() string

	//Login logs a user in a specific application using the specified credentials and authentication method.
	//The authentication method must be one of the supported for the application.
	//	Input:
	//		ipAddress (string): Client's IP address
	//		deviceType (string): "mobile" or "web" or "desktop" etc
	//		deviceOS (*string): Device OS
	//		deviceID (*string): Device ID
	//		authenticationType (string): Name of the authentication method for provided creds (eg. "password", "code", "illinois_oidc")
	//		creds (string): Credentials/JSON encoded credential structure defined for the specified auth type
	//		apiKey (string): API key to validate the specified app
	//		appTypeIdentifier (string): identifier of the app type/client that the user is logging in from
	//		orgID (string): ID of the organization that the user is logging in
	//		params (string): JSON encoded params defined by specified auth type
	//      clientVersion(*string): Most recent client version
	//		profile (Profile): Account profile
	//		preferences (map): Account preferences
	//		accountIdentifierID (*string): UUID of account identifier, meant to be used after using SignInOptions
	//		admin (bool): Is this an admin login?
	//		l (*logs.Log): Log object pointer for request
	//	Returns:
	//		Response parameters (map): any messages or parameters to send in response when requiring identifier verification and/or NOT logging in the user
	//		Login session (*LoginSession): Signed ROKWIRE access token to be used to authorize future requests
	//			Access token (string): Signed ROKWIRE access token to be used to authorize future requests
	//			Refresh Token (string): Refresh token that can be sent to refresh the access token once it expires
	//			AccountAuthType (AccountAuthType): AccountAuthType object for authenticated user
	//			Params (interface{}): authType-specific set of parameters passed back to client
	//			State (string): login state used if account is enrolled in MFA
	//		MFA types ([]model.MFAType): list of MFA types account is enrolled in
	Login(ipAddress string, deviceType string, deviceOS *string, deviceID *string, authenticationType string, creds string, apiKey string,
		appTypeIdentifier string, orgID string, params string, clientVersion *string, profile model.Profile, privacy model.Privacy, preferences map[string]interface{},
		accountIdentifierID *string, admin bool, l *logs.Log) (map[string]interface{}, *model.LoginSession, []model.MFAType, error)

	//Logout logouts an account from app/org
	//	Input:
	//		allSessions (bool): If to remove the current session only or all sessions for the app/org for the account
	Logout(appID string, orgID string, currentAccountID string, sessionID string, allSessions bool, l *logs.Log) error

	//AccountExists checks if a user is already registered
	//	Input:
	//		userIdentifier (string): User identifier for the specified auth type
	//		apiKey (string): API key to validate the specified app
	//		appTypeIdentifier (string): identifier of the app type/client that the user is logging in from
	//		orgID (string): ID of the organization that the user is logging in
	//	Returns:
	//		accountExisted (bool): valid when error is nil
	AccountExists(identifierJSON string, apiKey string, appTypeIdentifier string, orgID string, authenticationType *string, userIdentifier *string) (bool, error)

	//CanSignIn checks if a user can sign in
	//	Input:
	//		userIdentifier (string): User identifier for the specified auth type
	//		apiKey (string): API key to validate the specified app
	//		appTypeIdentifier (string): identifier of the app type/client being used
	//		orgID (string): ID of the organization being used
	//	Returns:
	//		canSignIn (bool): valid when error is nil
	CanSignIn(identifierJSON string, apiKey string, appTypeIdentifier string, orgID string, authenticationType *string, userIdentifier *string) (bool, error)

	//CanLink checks if a user can link a new auth type
	//	Input:
	//		userIdentifier (string): User identifier for the specified auth type
	//		apiKey (string): API key to validate the specified app
	//		appTypeIdentifier (string): identifier of the app type/client being used
	//		orgID (string): ID of the organization being used
	//	Returns:
	//		canLink (bool): valid when error is nil
	CanLink(identifierJSON string, apiKey string, appTypeIdentifier string, orgID string, authenticationType *string, userIdentifier *string) (bool, error)

	//SignInOptions returns the identifiers and auth types that may be used to sign in to an account
	//	Input:
	//		userIdentifier (string): User identifier for the specified auth type
	//		apiKey (string): API key to validate the specified app
	//		appTypeIdentifier (string): identifier of the app type/client being used
	//		orgID (string): ID of the organization being used
	//	Returns:
	//		identifiers ([]model.AccountIdentifier): account identifiers that may be used for sign-in
	//		authTypes ([]model.AccountAuthType): account auth types that may be used for sign-in
	SignInOptions(identifierJSON string, apiKey string, appTypeIdentifier string, orgID string, authenticationType *string, userIdentifier *string, l *logs.Log) ([]model.AccountIdentifier, []model.AccountAuthType, error)

	//Refresh refreshes an access token using a refresh token
	//	Input:
	//		refreshToken (string): Refresh token
	//		apiKey (string): API key to validate the specified app
	//      clientVersion(*string): Most recent client version
	//		l (*logs.Log): Log object pointer for request
	//	Returns:
	//		Login session (*LoginSession): Signed ROKWIRE access token to be used to authorize future requests
	//			Access token (string): Signed ROKWIRE access token to be used to authorize future requests
	//			Refresh Token (string): Refresh token that can be sent to refresh the access token once it expires
	//			Params (interface{}): authType-specific set of parameters passed back to client
	Refresh(refreshToken string, apiKey string, clientVersion *string, l *logs.Log) (*model.LoginSession, error)

	//GetLoginURL returns a pre-formatted login url for SSO providers
	//	Input:
	//		authType (string): Name of the authentication method for provided creds (eg. "illinois_oidc")
	//		appTypeIdentifier (string): Identifier of the app type/client that the user is logging in from
	//		orgID (string): ID of the organization that the user is logging in
	//		redirectURI (string): Registered redirect URI where client will receive response
	//		apiKey (string): API key to validate the specified app
	//		l (*loglib.Log): Log object pointer for request
	//	Returns:
	//		Login URL (string): SSO provider login URL to be launched in a browser
	//		Params (map[string]interface{}): Params to be sent in subsequent request (if necessary)
	GetLoginURL(authType string, appTypeIdentifier string, orgID string, redirectURI string, apiKey string, l *logs.Log) (string, map[string]interface{}, error)

	//LoginMFA verifies a code sent by a user as a final login step for enrolled accounts.
	//The MFA type must be one of the supported for the application.
	//	Input:
	//		apiKey (string): API key to validate the specified app
	//		accountID (string): ID of account user is trying to access
	//		sessionID (string): ID of login session generated during login
	//		identifier (string): Email, phone, or TOTP device name
	//		mfaType (string): Type of MFA code sent
	//		mfaCode (string): Code that must be verified
	//		state (string): Variable used to verify user has already passed credentials check
	//		l (*logs.Log): Log object pointer for request
	//	Returns:
	//		Message (*string): message
	//		Login session (*LoginSession): Signed ROKWIRE access token to be used to authorize future requests
	//			Access token (string): Signed ROKWIRE access token to be used to authorize future requests
	//			Refresh Token (string): Refresh token that can be sent to refresh the access token once it expires
	//			AccountAuthType (AccountAuthType): AccountAuthType object for authenticated user
	LoginMFA(apiKey string, accountID string, sessionID string, identifier string, mfaType string, mfaCode string, state string, l *logs.Log) (*string, *model.LoginSession, error)

	//CreateAdminAccount creates an account for a new admin user
	CreateAdminAccount(authenticationType string, appID string, orgID string, identifierJSON string, profile model.Profile, privacy model.Privacy, permissions []string,
		roleIDs []string, groupIDs []string, scopes []string, creatorPermissions []string, clientVersion *string, l *logs.Log) (*model.Account, map[string]interface{}, error)

	//UpdateAdminAccount updates an existing user's account with new permissions, roles, and groups
	UpdateAdminAccount(authenticationType string, appID string, orgID string, identifierJSON string, permissions []string, roleIDs []string,
		groupIDs []string, scopes []string, updaterPermissions []string, l *logs.Log) (*model.Account, map[string]interface{}, error)

	//CreateAnonymousAccount creates a new anonymous account
	CreateAnonymousAccount(context storage.TransactionContext, appID string, orgID string, anonymousID string, preferences map[string]interface{},
		systemConfigs map[string]interface{}, skipExistsCheck bool, l *logs.Log) (*model.Account, error)

	//VerifyIdentifier verifies credential (checks the verification code in the credentials collection)
	VerifyIdentifier(id string, verification string, l *logs.Log) (*model.AccountIdentifier, error)

	//SendVerifyIdentifier sends verification code to identifier
	SendVerifyIdentifier(appTypeIdentifier string, orgID string, apiKey string, identifierJSON string, l *logs.Log) error

	//UpdateCredential updates the credential object with the new value
	//	Input:
	//		accountID: id of the associated account to reset
	//		accountAuthTypeID (string): id of the AccountAuthType
	//		params: specific params for the different auth types
	//	Returns:
	//		error: if any
	UpdateCredential(accountID string, accountAuthTypeID string, params string, l *logs.Log) error

	//ForgotCredential initiate forgot credential process (generates a reset link and sends to the given identifier for email auth type)
	//	Input:
	//		authenticationType (string): Name of the authentication method for provided creds (eg. "password")
	//		identifierJSON (string): JSON string of the user's identifier and the identifier code
	//		appTypeIdentifier (string): Identifier of the app type/client that the user is logging in from
	//		orgID (string): ID of the organization that the user is logging in
	//		apiKey (string): API key to validate the specified app
	//	Returns:
	//		error: if any
	ForgotCredential(authenticationType string, identifierJSON string, appTypeIdentifier string, orgID string, apiKey string, l *logs.Log) error

	//ResetForgotCredential resets forgot credential
	//	Input:
	//		credsID: id of the credential object
	//		resetCode: code from the reset link
	//		params: specific params for the different auth types
	//	Returns:
	//		error: if any
	ResetForgotCredential(credsID string, resetCode string, params string, l *logs.Log) error

	//VerifyMFA verifies a code sent by a user as a final MFA enrollment step.
	//The MFA type must be one of the supported for the application.
	//	Input:
	//		accountID (string): ID of account for which user is trying to verify MFA
	//		identifier (string): Email, phone, or TOTP device name
	//		mfaType (string): Type of MFA code sent
	//		mfaCode (string): Code that must be verified
	//	Returns:
	//		Message (*string): message
	//		Recovery codes ([]string): List of account recovery codes returned if enrolling in MFA for first time
	VerifyMFA(accountID string, identifier string, mfaType string, mfaCode string) (*string, []string, error)

	//GetMFATypes gets all MFA types set up for an account
	//	Input:
	//		accountID (string): Account ID to find MFA types
	//	Returns:
	//		MFA Types ([]model.MFAType): MFA information for all enrolled types
	GetMFATypes(accountID string) ([]model.MFAType, error)

	//AddMFAType adds a form of MFA to an account
	//	Input:
	//		accountID (string): Account ID to add MFA
	//		identifier (string): Email, phone, or TOTP device name
	//		mfaType (string): Type of MFA to be added
	//	Returns:
	//		MFA Type (*model.MFAType): MFA information for the specified type
	AddMFAType(accountID string, identifier string, mfaType string) (*model.MFAType, error)

	//RemoveMFAType removes a form of MFA from an account
	//	Input:
	//		accountID (string): Account ID to remove MFA
	//		identifier (string): Email, phone, or TOTP device name
	//		mfaType (string): Type of MFA to remove
	RemoveMFAType(accountID string, identifier string, mfaType string) error

	//GetServiceAccountParams returns a list of app, org pairs a service account has access to
	GetServiceAccountParams(accountID string, firstParty bool, r *sigauth.Request, l *logs.Log) ([]model.AppOrgPair, error)

	//GetServiceAccessToken returns an access token for a non-human client
	GetServiceAccessToken(firstParty bool, r *sigauth.Request, l *logs.Log) (string, error)

	//GetAllServiceAccessTokens returns an access token for each app, org pair a service account has access to
	GetAllServiceAccessTokens(firstParty bool, r *sigauth.Request, l *logs.Log) (map[model.AppOrgPair]string, error)

	//GetServiceAccounts gets all service accounts matching a search
	GetServiceAccounts(params map[string]interface{}) ([]model.ServiceAccount, error)

	//RegisterServiceAccount registers a service account
	RegisterServiceAccount(accountID *string, fromAppID *string, fromOrgID *string, name *string, appID string, orgID string, permissions *[]string, scopes []authorization.Scope,
		firstParty *bool, creds []model.ServiceAccountCredential, assignerPermissions []string, l *logs.Log) (*model.ServiceAccount, error)

	//DeregisterServiceAccount deregisters a service account
	DeregisterServiceAccount(accountID string) error

	//GetServiceAccountInstance gets a service account instance
	GetServiceAccountInstance(accountID string, appID string, orgID string) (*model.ServiceAccount, error)

	//UpdateServiceAccountInstance updates a service account instance
	UpdateServiceAccountInstance(id string, appID string, orgID string, name *string, permissions *[]string, scopes []authorization.Scope, assignerPermissions []string) (*model.ServiceAccount, error)

	//DeregisterServiceAccountInstance deregisters a service account instance
	DeregisterServiceAccountInstance(id string, appID string, orgID string) error

	//AddServiceAccountCredential adds a credential to a service account
	AddServiceAccountCredential(accountID string, creds *model.ServiceAccountCredential, l *logs.Log) (*model.ServiceAccountCredential, error)

	//RemoveServiceAccountCredential removes a credential from a service account
	RemoveServiceAccountCredential(accountID string, credID string) error

	//AuthorizeService returns a scoped token for the specified service and the service registration record if authorized or
	//	the service registration record if not. Passing "approvedScopes" will update the service authorization for this user and
	//	return a scoped access token which reflects this change.
	//	Input:
	//		claims (tokenauth.Claims): Claims from un-scoped user access token
	//		serviceID (string): ID of the service to be authorized
	//		approvedScopes ([]string): list of scope strings to be approved
	//		l (*logs.Log): Log object pointer for request
	//	Returns:
	//		Access token (string): Signed scoped access token to be used to authorize requests to the specified service
	//		Approved Scopes ([]authorization.Scope): The approved scopes included in the provided token
	//		Service reg (*model.ServiceReg): The service registration record for the requested service
	AuthorizeService(claims tokenauth.Claims, serviceID string, approvedScopes []authorization.Scope, l *logs.Log) (string, []authorization.Scope, *model.ServiceRegistration, error)

	//LinkAccountAuthType links new credentials to an existing account.
	//The authentication method must be one of the supported for the application.
	//	Input:
	//		orgID (string): Org id
	//		appID (string): App id
	//		accountID (string): ID of the account to link the creds to
	//		authenticationType (string): Name of the authentication method for provided creds (eg. "password", "webauthn", "illinois_oidc")
	//		appTypeIdentifier (string): Identifier of the app type/client that the user is logging in from
	//		creds (string): Credentials/JSON encoded credential structure defined for the specified auth type
	//		params (string): JSON encoded params defined by specified auth type
	//		l (*logs.Log): Log object pointer for request
	//	Returns:
	//		message (*string): response message
	//		account (*model.Account): account data after the operation
	LinkAccountAuthType(orgID string, appID string, accountID string, authenticationType string, appTypeIdentifier string, creds string, params string, l *logs.Log) (*string, *model.Account, error)

	//UnlinkAccountAuthType unlinks credentials from an existing account.
	//The authentication method must be one of the supported for the application.
	//	Input:
	//		accountID (string): ID of the account to unlink creds from
	//		accountAuthTypeID (*string): Account auth type to unlink
	//		authenticationType (*string): Name of the authentication method of account auth type to unlink
	//		identifier (*string): Identifier to unlink
	//		l (*logs.Log): Log object pointer for request
	//	Returns:
	//		account (*model.Account): account data after the operation
	UnlinkAccountAuthType(accountID string, accountAuthTypeID *string, authenticationType *string, identifier *string, admin bool, l *logs.Log) (*model.Account, error)

	LinkAccountIdentifier(accountID string, identifierJSON string, admin bool, l *logs.Log) (*string, *model.Account, error)

	UnlinkAccountIdentifier(accountID string, accountIdentifierID string, admin bool, l *logs.Log) (*model.Account, error)

	//InitializeSystemAccount initializes the first system account
	InitializeSystemAccount(context storage.TransactionContext, authType model.AuthType, appOrg model.ApplicationOrganization, allSystemPermission string, email string, password string, clientVersion string, l *logs.Log) (string, error)

	//GrantAccountPermissions grants new permissions to an account after validating the assigner has required permissions
	//GrantAccountPermissions(context storage.TransactionContext, account *model.Account, permissionNames []string, assignerPermissions []string) error

	//CheckPermissions loads permissions by names from storage and checks that they are assignable and valid for the given appOrgs or revocable
	CheckPermissions(context storage.TransactionContext, appOrgs []model.ApplicationOrganization, permissionNames []string, assignerPermissions []string, revoke bool) ([]model.Permission, error)

	//GrantAccountRoles grants new roles to an account after validating the assigner has required permissions
	GrantAccountRoles(context storage.TransactionContext, account *model.Account, roleIDs []string, assignerPermissions []string) error

	//CheckRoles loads appOrg roles by IDs from storage and checks that they are assignable or revocable
	CheckRoles(context storage.TransactionContext, appOrg *model.ApplicationOrganization, roleIDs []string, assignerPermissions []string, revoke bool) ([]model.AppOrgRole, error)

	//GrantAccountGroups grants new groups to an account after validating the assigner has required permissions
	GrantAccountGroups(context storage.TransactionContext, account *model.Account, groupIDs []string, assignerPermissions []string) error

	//CheckGroups loads appOrg groups by IDs from storage and checks that they are assignable or revocable
	CheckGroups(context storage.TransactionContext, appOrg *model.ApplicationOrganization, groupIDs []string, assignerPermissions []string, revoke bool) ([]model.AppOrgGroup, error)

	//DeleteAccount deletes an account for the given id
	DeleteAccount(id string, apps []string) error

	//GetAdminToken returns an admin token for the specified application and organization
	GetAdminToken(claims tokenauth.Claims, appID string, orgID string, l *logs.Log) (string, error)

	// EncryptSecrets JSON encodes and encrypts the given plain secrets
	EncryptSecrets(secrets map[string]interface{}) (map[string]interface{}, error)

	// DecryptSecrets decrypts and JSON decodes the given encrypted secrets
	DecryptSecrets(secrets map[string]interface{}) (map[string]interface{}, error)

	//GetAuthKeySet generates a JSON Web Key Set for auth service registration
	GetAuthKeySet() (jwk.Set, error)

	//GetServiceRegistrations retrieves all service registrations
	GetServiceRegistrations(serviceIDs []string) []model.ServiceRegistration

	//RegisterService creates a new service registration
	RegisterService(reg *model.ServiceRegistration) error

	//UpdateServiceRegistration updates an existing service registration
	UpdateServiceRegistration(reg *model.ServiceRegistration) error

	//DeregisterService deletes an existing service registration
	DeregisterService(serviceID string) error

	//GetApplicationAPIKeys finds and returns the API keys for an application
	GetApplicationAPIKeys(appID string) ([]model.APIKey, error)

	//GetAPIKey finds and returns an API key
	GetAPIKey(ID string) (*model.APIKey, error)

	//CreateAPIKey creates a new API key
	CreateAPIKey(apiKey model.APIKey) (*model.APIKey, error)

	//UpdateAPIKey updates an existing API key
	UpdateAPIKey(apiKey model.APIKey) error

	//DeleteAPIKey deletes an API key
	DeleteAPIKey(ID string) error

	//ValidateAPIKey validates the given API key for the given app ID
	ValidateAPIKey(appID string, apiKey string) error
}

// Storage interface to communicate with the storage
type Storage interface {
	RegisterStorageListener(storageListener storage.Listener)

	PerformTransaction(func(context storage.TransactionContext) error) error

	//Configs
	FindConfig(configType string, appID string, orgID string) (*model.Config, error)

	//AuthTypes
	FindAuthType(codeOrID string) (*model.AuthType, error)

	//LoginsSessions
	InsertLoginSession(context storage.TransactionContext, session model.LoginSession) error
	FindLoginSessions(context storage.TransactionContext, identifier string) ([]model.LoginSession, error)
	FindLoginSession(refreshToken string) (*model.LoginSession, error)
	FindAndUpdateLoginSession(context storage.TransactionContext, id string) (*model.LoginSession, error)
	UpdateLoginSession(context storage.TransactionContext, loginSession model.LoginSession) error
	DeleteLoginSession(context storage.TransactionContext, id string) error
	DeleteLoginSessionsByIDs(context storage.TransactionContext, ids []string) error
	DeleteLoginSessionsByIdentifier(context storage.TransactionContext, identifier string) error

	//LoginsSessions - predefined queries for manage deletion logic
	DeleteMFAExpiredSessions() error
	FindSessionsLazy(appID string, orgID string) ([]model.LoginSession, error)
	///

	//LoginStates
	FindLoginState(appID string, orgID string, accountID *string, stateParams map[string]interface{}) (*model.LoginState, error)
	InsertLoginState(loginState model.LoginState) error
	DeleteLoginState(context storage.TransactionContext, id string) error

	//Accounts
	FindAccountByOrgAndIdentifier(context storage.TransactionContext, orgID string, code string, identifier string, currentAppOrgID string) (*model.Account, error)
	FindAccount(context storage.TransactionContext, appOrgID string, code string, identifier string) (*model.Account, error)
	FindAccountByID(context storage.TransactionContext, cOrgID string, cAppID string, id string) (*model.Account, error)
	FindAccountsByUsername(context storage.TransactionContext, appOrg *model.ApplicationOrganization, username string) ([]model.Account, error)
	InsertAccount(context storage.TransactionContext, account model.Account) (*model.Account, error)
	SaveAccount(context storage.TransactionContext, account *model.Account) error
	DeleteAccount(context storage.TransactionContext, id string) error
	UpdateAccountUsageInfo(context storage.TransactionContext, accountID string, updateLoginTime bool, updateAccessTokenTime bool, clientVersion *string) error
	DeleteOrgAppsMemberships(context storage.TransactionContext, accountID string, membershipsIDs []string) error

	//ServiceAccounts
	FindServiceAccount(context storage.TransactionContext, accountID string, appID string, orgID string) (*model.ServiceAccount, error)
	FindServiceAccounts(params map[string]interface{}) ([]model.ServiceAccount, error)
	InsertServiceAccount(account *model.ServiceAccount) error
	UpdateServiceAccount(context storage.TransactionContext, account *model.ServiceAccount) (*model.ServiceAccount, error)
	DeleteServiceAccount(accountID string, appID string, orgID string) error
	DeleteServiceAccounts(accountID string) error

	//ServiceAccountCredentials
	InsertServiceAccountCredential(accountID string, creds *model.ServiceAccountCredential) error
	DeleteServiceAccountCredential(accountID string, credID string) error

	//AccountAuthTypes
	FindAccountByAuthTypeID(context storage.TransactionContext, id string, currentAppOrgID *string) (*model.Account, error)
	FindAccountByCredentialID(context storage.TransactionContext, id string) (*model.Account, error)
	InsertAccountAuthType(context storage.TransactionContext, item model.AccountAuthType) error
	UpdateAccountAuthType(context storage.TransactionContext, item model.AccountAuthType) error
	DeleteAccountAuthType(context storage.TransactionContext, item model.AccountAuthType) error

	//AccountIdentifiers
	FindAccountByIdentifierID(context storage.TransactionContext, id string) (*model.Account, error)
	InsertAccountIdentifier(context storage.TransactionContext, item model.AccountIdentifier) error
	UpdateAccountIdentifier(context storage.TransactionContext, item model.AccountIdentifier) error
	UpdateAccountIdentifiers(context storage.TransactionContext, accountID string, items []model.AccountIdentifier) error
	DeleteAccountIdentifier(context storage.TransactionContext, item model.AccountIdentifier) error
	DeleteExternalAccountIdentifiers(context storage.TransactionContext, aat model.AccountAuthType) error

	//Applications
	FindApplication(context storage.TransactionContext, ID string) (*model.Application, error)

	//Organizations
	FindOrganization(id string) (*model.Organization, error)

	//Credentials
	InsertCredential(context storage.TransactionContext, creds *model.Credential) error
	FindCredential(context storage.TransactionContext, ID string) (*model.Credential, error)
	FindCredentials(context storage.TransactionContext, ids []string) ([]model.Credential, error)
	UpdateCredential(context storage.TransactionContext, creds *model.Credential) error
	UpdateCredentialValue(context storage.TransactionContext, ID string, value map[string]interface{}) error
	DeleteCredential(context storage.TransactionContext, ID string) error

	//MFA
	FindMFAType(context storage.TransactionContext, accountID string, identifier string, mfaType string) (*model.MFAType, error)
	FindMFATypes(accountID string) ([]model.MFAType, error)
	InsertMFAType(context storage.TransactionContext, mfa *model.MFAType, accountID string) error
	UpdateMFAType(context storage.TransactionContext, mfa *model.MFAType, accountID string) error
	DeleteMFAType(context storage.TransactionContext, accountID string, identifier string, mfaType string) error

	//ServiceRegs
	FindServiceRegs(serviceIDs []string) []model.ServiceRegistration
	FindServiceReg(serviceID string) (*model.ServiceRegistration, error)
	InsertServiceReg(reg *model.ServiceRegistration) error
	UpdateServiceReg(reg *model.ServiceRegistration) error
	SaveServiceReg(reg *model.ServiceRegistration, immediateCache bool) error
	DeleteServiceReg(serviceID string) error
	MigrateServiceRegs() error

	//IdentityProviders
	LoadIdentityProviders() ([]model.IdentityProvider, error)

	//ServiceAuthorizations
	FindServiceAuthorization(userID string, orgID string) (*model.ServiceAuthorization, error)
	SaveServiceAuthorization(authorization *model.ServiceAuthorization) error
	DeleteServiceAuthorization(userID string, orgID string) error

	//Keys
	FindKey(name string) (*model.Key, error)
	InsertKey(key model.Key) error
	UpdateKey(key model.Key) error

	//APIKeys
	LoadAPIKeys() ([]model.APIKey, error)
	InsertAPIKey(context storage.TransactionContext, apiKey model.APIKey) (*model.APIKey, error)
	UpdateAPIKey(apiKey model.APIKey) error
	DeleteAPIKey(ID string) error

	//ApplicationTypes
	FindApplicationType(id string) (*model.ApplicationType, error)

	//ApplicationsOrganizations
	FindApplicationsOrganizations() ([]model.ApplicationOrganization, error)
	FindApplicationOrganizations(appID *string, orgID *string) ([]model.ApplicationOrganization, error)
	FindApplicationOrganization(appID string, orgID string) (*model.ApplicationOrganization, error)

	//Device
	FindDevice(context storage.TransactionContext, deviceID *string, accountID string) (*model.Device, error)
	InsertDevice(context storage.TransactionContext, device model.Device) (*model.Device, error)
	DeleteDevice(context storage.TransactionContext, id string) error

	//Permissions
	FindPermissions(context storage.TransactionContext, ids []string) ([]model.Permission, error)
	FindPermissionsByName(context storage.TransactionContext, names []string) ([]model.Permission, error)
	InsertAccountPermissions(context storage.TransactionContext, accountID string, appOrgID string, permissions []model.Permission) error
	UpdateAccountPermissions(context storage.TransactionContext, accountID string, appOrgID string, permissions []model.Permission) error

	//ApplicationRoles
	FindAppOrgRolesByIDs(context storage.TransactionContext, ids []string, appOrgID string) ([]model.AppOrgRole, error)
	//AccountRoles
	UpdateAccountRoles(context storage.TransactionContext, accountID string, appOrgID string, roles []model.AccountRole) error
	InsertAccountRoles(context storage.TransactionContext, accountID string, appOrgID string, roles []model.AccountRole) error

	UpdateAccountScopes(context storage.TransactionContext, accountID string, scopes []string) error

	//ApplicationGroups
	FindAppOrgGroupsByIDs(context storage.TransactionContext, ids []string, appOrgID string) ([]model.AppOrgGroup, error)
	//AccountGroups
	UpdateAccountGroups(context storage.TransactionContext, accountID string, appOrgID string, groups []model.AccountGroup) error
	InsertAccountGroups(context storage.TransactionContext, accountID string, appOrgID string, groups []model.AccountGroup) error
}

// ProfileBuildingBlock is used by auth to communicate with the profile building block.
type ProfileBuildingBlock interface {
	GetProfileBBData(queryParams map[string]string, l *logs.Log) (*model.Profile, map[string]interface{}, error)
}

// IdentityBuildingBlock is used by auth to communicate with the identity building block.
type IdentityBuildingBlock interface {
	GetUserProfile(baseURL string, externalUser model.ExternalSystemUser, externalAccessToken string, l *logs.Log) (*model.Profile, error)
}

// Emailer is used by core to send emails
type Emailer interface {
	Send(toEmail string, subject string, body string, attachmentFilename *string) error
}

// PhoneVerifier is used by core to verify phone numbers
type PhoneVerifier interface {
	StartVerification(phone string, data url.Values) error
	CheckVerification(phone string, data url.Values) error
}
