package auth

import (
	"core-building-block/core/model"
	"core-building-block/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rokmetro/logging-library/errors"
	"github.com/rokmetro/logging-library/logs"
	"github.com/rokmetro/logging-library/logutils"
	"golang.org/x/crypto/bcrypt"
)

const (
	typeTime        logutils.MessageDataType = "time.Time"
	authTypeEmail   string                   = "email"
	typeEmailCreds  logutils.MessageDataType = "email creds"
	typeEmailParams logutils.MessageDataType = "email params"
)

//enailCreds represents the creds struct for email auth
type emailCreds struct {
	Email              string    `json:"email" bson:"email" validate:"required"`
	Password           string    `json:"password" bson:"password"`
	VerificationCode   string    `json:"verification_code" bson:"verification_code"`
	VerificationExpiry time.Time `json:"verification_expiry" bson:"verification_expiry"`
}

// Email implementation of authType
type emailAuthImpl struct {
	auth     *Auth
	authType string
}

func (a *emailAuthImpl) applySignUp(authType model.AuthType, appType model.ApplicationType, appOrg model.ApplicationOrganization, creds string, params string, l *logs.Log) (*string, map[string]interface{}, error) {
	type signUpEmailParams struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	var sEmailParams signUpEmailParams
	err := json.Unmarshal([]byte(params), &sEmailParams)
	if err != nil {
		return nil, nil, errors.WrapErrorAction("error getting sign_up email params", "", nil, err)
	}
	email := sEmailParams.Email
	password := sEmailParams.Password
	confirmPassword := sEmailParams.ConfirmPassword
	if len(email) == 0 || len(password) == 0 || len(confirmPassword) == 0 {
		return nil, nil, errors.WrapErrorAction("bad params data", "", nil, err)
	}
	//check if the passwrod matches with the confirm password one
	if password != confirmPassword {
		return nil, nil, errors.WrapErrorAction("passwords fields do not match", "", nil, err)
	}

	//password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionCompute, model.TypeAuthCred, nil, errors.New("failed to generate hash from password"))
	}

	//verification code
	code, err := utils.GenerateRandomString(64)
	if err != nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionCompute, model.TypeAuthCred, nil, errors.New("failed to generate random string for verify code"))

	}

	//send verification code
	if err = a.sendVerificationCode(email, code, appOrg.Application.ID, appOrg.Organization.ID); err != nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAuthCred, nil, errors.New("failed to send verification email to user"))
	}

	emailCredValue := emailCreds{Email: email, Password: string(hashedPassword), VerificationCode: code, VerificationExpiry: time.Now().Add(time.Hour * 24)}
	emailCredValueMap, err := emailCredsToMap(&emailCredValue)
	if err != nil {
		return nil, nil, errors.WrapErrorAction("failed email params to map", "", nil, err)
	}

	return &email, emailCredValueMap, nil
}

// check(creds string, orgID string, appID string, params string, l *logs.Log) (*model.UserAuth, error)
func (a *emailAuthImpl) checkCredentials(accountAuthType *model.AccountAuthType, creds string, params string, appOrg model.ApplicationOrganization, l *logs.Log) (*string, map[string]interface{}, bool, error) {
	//appID := appOrg.Application.ID
	//orgID := appOrg.Organization.ID

	/*
		var requestCreds emailCreds
		err := json.Unmarshal([]byte(creds), &requestCreds)
		if err != nil {
			return nil, nil, false, errors.WrapErrorAction(logutils.ActionUnmarshal, typeEmailCreds, logutils.StringArgs("request"), err)
		}

		newUser := false
			var requestParams emailParams
			err = json.Unmarshal([]byte(params), &requestParams)
			if err == nil {
				newUser = requestParams.NewUser
			}

			var storedCreds *emailCreds
			if accountAuthType != nil && accountAuthType.Credential != nil {
				storedCreds, err = mapToEmailCreds(accountAuthType.Credential.Value)
				if err != nil {
					return nil, nil, false, errors.WrapErrorAction(logutils.ActionUnmarshal, typeEmailCreds, logutils.StringArgs("stored"), err)
				}
			}
	*/
	//Handle sign up
	/*	if accountAuthType == nil {
		if !newUser {
			return nil, nil, false, errors.New("no account found newUser flag must be set")
		}

		newCreds, err := a.handleSignup(&requestCreds, storedCreds, appID, orgID)
		if err != nil {
			return nil, nil, false, err
		}
		newCredsMap, err := emailCredsToMap(newCreds)
		if err != nil {
			return nil, nil, false, err
		}
		return &newCreds.Email, newCredsMap, false, nil
	} */
	/*
		if newUser {
			return nil, nil, false, errors.Newf("account already exists for email: %s", requestCreds.Email)
		}
		if err = a.handleSignin(&requestCreds, storedCreds); err != nil {
			return nil, nil, false, errors.WrapErrorAction(logutils.ActionValidate, typeEmailCreds, nil, err)
		}
		return &storedCreds.Email, accountAuthType.Credential.Value, accountAuthType.Credential.Verified, nil */
	return nil, nil, false, nil
}

func (a *emailAuthImpl) handleSignin(requestCreds *emailCreds, storageCreds *emailCreds) error {
	if storageCreds == nil {
		return errors.ErrorData(logutils.StatusMissing, model.TypeAuthCred, logutils.StringArgs("stored"))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storageCreds.Password), []byte(requestCreds.Password)); err != nil {
		return errors.WrapErrorAction(logutils.ActionValidate, model.TypeAuthCred, nil, err)
	}
	return nil
}

func (a *emailAuthImpl) sendVerificationCode(email string, verificationCode string, appID string, orgID string) error {
	verificationLink := a.auth.host + fmt.Sprintf("/auth/verify/%v/%v/%v/%v/%v", a.authType, email, verificationCode, appID, orgID)
	return a.auth.sendEmail(email, "Verify your email", "Please click the link below to verify your email:\n"+verificationLink, "")
}

//TODO: To be used in password reset flow
// func (a *emailAuthImpl) sendPasswordReset(email string, password string) error {
// 	return a.auth.SendEmail(email, "Password Reset", "Your temporary password is "+password, "")
// }

func (a *emailAuthImpl) verify(credential *model.Credential, verification string, l *logs.Log) (map[string]interface{}, error) {
	credBytes, err := json.Marshal(credential.Value)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionMarshal, typeEmailCreds, nil, err)
	}

	var creds *emailCreds
	err = json.Unmarshal(credBytes, &creds)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionUnmarshal, typeEmailCreds, nil, err)
	}
	err = a.compareVerifyCode(creds.VerificationCode, verification, creds.VerificationExpiry, l)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, model.TypeAuthCred, &logutils.FieldArgs{"verification_code": verification}, errors.New("invalid verification code"))
	}

	//Update verification data
	creds.VerificationCode = ""
	creds.VerificationExpiry = time.Time{}
	credsMap, err := emailCredsToMap(creds)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCast, typeEmailCreds, nil, err)
	}

	return credsMap, nil
}

func (a *emailAuthImpl) compareVerifyCode(credCode string, requestCode string, expiryTime time.Time, l *logs.Log) error {
	if expiryTime.Before(time.Now()) {
		return errors.WrapErrorAction(logutils.ActionValidate, typeTime, nil, errors.New("verification code has expired"))
	}

	if credCode != requestCode {
		return errors.WrapErrorAction(logutils.ActionValidate, typeTime, nil, errors.New("Invalid verification code"))
	}
	return nil

}
func (a *emailAuthImpl) userExist(authType model.AuthType, appType model.ApplicationType, appOrg model.ApplicationOrganization, creds string, l *logs.Log) (*model.Account, *model.AccountAuthType, error) {
	appID := appOrg.Application.ID
	orgID := appOrg.Organization.ID
	authTypeID := authType.ID

	var requestCreds emailCreds
	err := json.Unmarshal([]byte(creds), &requestCreds)
	if err != nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionUnmarshal, typeEmailCreds, logutils.StringArgs("request"), err)
	}

	account, err := a.auth.storage.FindAccount(appID, orgID, authTypeID, requestCreds.Email)
	if err != nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAccount, nil, err) //TODO add args..
	}

	if account == nil {
		return nil, nil, nil
	}

	accountAuthType, err := a.auth.findAccountAuthType(account, authTypeID, requestCreds.Email)
	if accountAuthType == nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAccountAuthType, nil, err) //TODO add args..
	}

	return account, accountAuthType, nil
}

func emailCredsToMap(creds *emailCreds) (map[string]interface{}, error) {
	credBytes, err := json.Marshal(creds)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionMarshal, model.TypeAuthCred, nil, err)
	}
	var credsMap map[string]interface{}
	err = json.Unmarshal(credBytes, &credsMap)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionUnmarshal, model.TypeAuthCred, nil, err)
	}
	return credsMap, nil
}

func mapToEmailCreds(credsMap map[string]interface{}) (*emailCreds, error) {
	credBytes, err := json.Marshal(credsMap)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionMarshal, typeEmailCreds, nil, err)
	}
	var creds emailCreds
	err = json.Unmarshal(credBytes, &creds)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionUnmarshal, typeEmailCreds, nil, err)
	}
	return &creds, nil
}

//initEmailAuth initializes and registers a new email auth instance
func initEmailAuth(auth *Auth) (*emailAuthImpl, error) {
	email := &emailAuthImpl{auth: auth, authType: authTypeEmail}

	err := auth.registerAuthType(email.authType, email)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionRegister, typeAuthType, nil, err)
	}

	return email, nil
}
