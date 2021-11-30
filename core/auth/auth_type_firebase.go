package auth

import (
	"core-building-block/core/model"

	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
)

const (
	//AuthTypeFirebase firebase auth type
	AuthTypeFirebase string = "firebase"
)

// Firebase implementation of authType
type firebaseAuthImpl struct {
	auth     *Auth
	authType string
}

func (a *firebaseAuthImpl) signUp(authType model.AuthType, appType model.ApplicationType, appOrg model.ApplicationOrganization, creds string, params string, newCredentialID string, l *logs.Log) (string, map[string]interface{}, error) {
	return "", nil, nil
}

func (a *firebaseAuthImpl) getUserIdentifier(creds string) (string, error) {
	return "", nil
}

func (a *firebaseAuthImpl) verifyCredential(credential *model.Credential, verification string, l *logs.Log) (map[string]interface{}, error) {
	return nil, errors.New(logutils.Unimplemented)
}

func (a *firebaseAuthImpl) sendVerifyCredential(credential *model.Credential, l *logs.Log) error {
	return nil
}

func (a *firebaseAuthImpl) restartCredentialVerification(credential *model.Credential, l *logs.Log) error {
	return nil
}

func (a *firebaseAuthImpl) isCredentialVerified(credential *model.Credential, l *logs.Log) (*bool, *bool, error) {
	return nil, nil, nil
}

func (a *firebaseAuthImpl) checkCredentials(accountAuthType model.AccountAuthType, creds string, l *logs.Log) (string, error) {
	return "", nil
}

func (a *firebaseAuthImpl) resetCredential(credential *model.Credential, resetCode *string, params string, l *logs.Log) (map[string]interface{}, error) {
	return nil, nil
}

func (a *firebaseAuthImpl) forgotCredential(credential *model.Credential, identifier string, l *logs.Log) (map[string]interface{}, error) {
	return nil, nil
}

//initFirebaseAuth initializes and registers a new Firebase auth instance
func initFirebaseAuth(auth *Auth) (*firebaseAuthImpl, error) {
	firebase := &firebaseAuthImpl{auth: auth, authType: AuthTypeFirebase}

	err := auth.registerAuthType(firebase.authType, firebase)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionRegister, typeAuthType, nil, err)
	}

	return firebase, nil
}