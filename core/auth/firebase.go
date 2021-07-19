package auth

import (
	"core-building-block/core/model"
	"errors"
)

// Firebase implementation of authType
type firebaseAuthImpl struct {
	auth *Auth
}

func (a *firebaseAuthImpl) check(creds string, params string) (*model.UserAuth, error) {
	//TODO: Implement
	return nil, errors.New("Unimplemented")
}

//initFirebaseAuth initializes and registers a new Firebase auth instance
func initFirebaseAuth(auth *Auth) (*firebaseAuthImpl, error) {
	firebase := &firebaseAuthImpl{auth: auth}

	err := auth.registerAuthType("firebase", firebase)
	if err != nil {
		return nil, err
	}

	return firebase, nil
}
