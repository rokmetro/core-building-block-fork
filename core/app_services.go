package core

import (
	"core-building-block/core/model"

	"github.com/rokmetro/logging-library/logs"
)

func (app *application) serGetPII(ID string) (*model.Pii, error) {
	return app.storage.FindPII(ID)
}

func (app *application) serUpdatePII(pii *model.Pii, ID string) error {
	return app.storage.UpdatePII(pii, ID)
}

func (app *application) serDeletePII(ID string) error {
	return app.storage.DeletePII(ID)
}

func (app *application) serGetAuthTest(l *logs.Log) string {
	return "Services - Auth - test"
}

func (app *application) serGetCommonTest(l *logs.Log) string {
	return "Services - Common - test"
}
