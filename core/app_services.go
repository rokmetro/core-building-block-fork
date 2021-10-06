package core

import (
	"core-building-block/core/model"

	"github.com/rokmetro/logging-library/errors"
	"github.com/rokmetro/logging-library/logs"
	"github.com/rokmetro/logging-library/logutils"
)

func (app *application) serGetProfile(accountID string) (*model.Profile, error) {
	//find the account
	account, err := app.storage.FindAccountByID(accountID)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAccount, nil, err)
	}

	//get the profile for the account
	profile := account.Profile
	return &profile, nil
}

func (app *application) serGetPreferences(accountID string) (*model.Account, map[string]interface{}, error) {
	//find the account
	account, err := app.storage.FindAccountByID(accountID)
	if err != nil {
		return nil, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAccount, nil, err)
	}

	//get the preferences for the account
	preferences := account.Preferences
	return nil, preferences, nil
}

func (app *application) serUpdateProfile(accountID string, profile *model.Profile) error {
	return app.storage.UpdateProfile(accountID, profile)
}

func (app *application) serUpdateAccountPreferences(id string, preferences map[string]interface{}) error {
	err := app.storage.UpdateAccountPreferences(id, preferences)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeAccount, nil, err)
	}
	return nil
}

func (app *application) serDeleteAccount(id string) error {
	return app.storage.DeleteAccount(id)
}

func (app *application) serGetAuthTest(l *logs.Log) string {
	return "Services - Auth - test"
}

func (app *application) serGetCommonTest(l *logs.Log) string {
	return "Services - Common - test"
}
