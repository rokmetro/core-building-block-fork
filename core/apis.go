package core

import (
	"core-building-block/core/auth"
	"core-building-block/core/model"

	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
)

//APIs exposes to the drivers adapters access to the core functionality
type APIs struct {
	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivers adapters
	Encryption     Encryption     //expose to the drivers adapters
	BBs            BBs            //expose to the drivers adapters
	System         System         //expose to the drivers adapters

	Auth auth.APIs //expose to the drivers auth

	app *application
}

//Start starts the core part of the application
func (c *APIs) Start() {
	c.app.start()
	c.Auth.Start()
}

//AddListener adds application listener
func (c *APIs) AddListener(listener ApplicationListener) {
	c.app.addListener(listener)
}

//GetVersion gives the service version
func (c *APIs) GetVersion() string {
	return c.app.version
}

//NewCoreAPIs creates new CoreAPIs
func NewCoreAPIs(env string, version string, build string, storage Storage, auth auth.APIs) *APIs {
	//add application instance
	listeners := []ApplicationListener{}
	application := application{env: env, version: version, build: build, storage: storage, listeners: listeners}

	//add coreAPIs instance
	servicesImpl := &servicesImpl{app: &application, auth: auth}
	administrationImpl := &administrationImpl{app: &application}
	encryptionImpl := &encryptionImpl{app: &application}
	bbsImpl := &bbsImpl{app: &application}
	systemImpl := &systemImpl{app: &application}

	//+ auth
	coreAPIs := APIs{Services: servicesImpl, Administration: administrationImpl, Encryption: encryptionImpl,
		BBs: bbsImpl, System: systemImpl, Auth: auth, app: &application}

	return &coreAPIs
}

///

//servicesImpl
type servicesImpl struct {
	auth auth.APIs
	app  *application
}

func (s *servicesImpl) SerDeleteAccount(id string) error {
	return s.app.serDeleteAccount(id)
}

func (s *servicesImpl) SerGetAccount(accountID string) (*model.Account, error) {
	return s.app.serGetAccount(accountID)
}

func (s *servicesImpl) SerGetProfile(accountID string) (*model.Profile, error) {
	return s.app.serGetProfile(accountID)
}

func (s *servicesImpl) SerGetPreferences(accountID string) (map[string]interface{}, error) {
	return s.app.serGetPreferences(accountID)
}

func (s *servicesImpl) SerUpdateProfile(accountID string, profile *model.Profile) error {
	return s.app.serUpdateProfile(accountID, profile)
}

func (s *servicesImpl) SerGetAuthTest(l *logs.Log) string {
	return s.app.serGetAuthTest(l)
}

func (s *servicesImpl) SerGetCommonTest(l *logs.Log) string {
	return s.app.serGetCommonTest(l)
}

func (s *servicesImpl) SerUpdateAccountPreferences(id string, preferences map[string]interface{}) error {
	return s.app.serUpdateAccountPreferences(id, preferences)
}

func (s *servicesImpl) SerGetAppConfig(appTypeIdentifier string, appOrgID string, versionNumbers model.VersionNumbers, apiKey string) (*model.ApplicationConfig, error) {
	//get the app type
	applicationType, err := s.app.storage.FindApplicationTypeByIdentifier(appTypeIdentifier)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeApplicationType, logutils.StringArgs(appTypeIdentifier), err)

	}
	if applicationType == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeApplicationType, logutils.StringArgs(appTypeIdentifier))
	}

	appID := applicationType.Application.ID

	err = s.auth.ValidateAPIKey(appID, apiKey)
	if err != nil {
		return nil, errors.WrapErrorData(logutils.StatusInvalid, model.TypeAPIKey, nil, err)
	}

	return s.app.serGetAppConfig(applicationType.ID, appOrgID, versionNumbers)
}

///

//administrationImpl

type administrationImpl struct {
	app *application
}

func (s *administrationImpl) AdmGetTest() string {
	return s.app.admGetTest()
}

func (s *administrationImpl) AdmGetTestModel() string {
	return s.app.admGetTestModel()
}

func (s *administrationImpl) AdmGetAccounts(appID string, orgID string, accountID *string, authTypeIdentifier *string) ([]model.Account, error) {
	return s.app.admGetAccounts(appID, orgID, accountID, authTypeIdentifier)
}

func (s *administrationImpl) AdmGetAccount(accountID string) (*model.Account, error) {
	return s.app.admGetAccount(accountID)
}

///

//encryptionImpl

type encryptionImpl struct {
	app *application
}

func (s *encryptionImpl) EncGetTest() string {
	return s.app.encGetTest()
}

///

//bbsImpl

type bbsImpl struct {
	app *application
}

func (s *bbsImpl) BBsGetTest() string {
	return s.app.bbsGetTest()
}

///

//systemImpl

type systemImpl struct {
	app *application
}

func (s *systemImpl) SysCreateGlobalConfig(setting string) (*model.GlobalConfig, error) {
	return s.app.sysCreateGlobalConfig(setting)
}

func (s *systemImpl) SysGetGlobalConfig() (*model.GlobalConfig, error) {
	return s.app.sysGetGlobalConfig()
}

func (s *systemImpl) SysUpdateGlobalConfig(setting string) error {
	return s.app.sysUpdateGlobalConfig(setting)
}

func (s *systemImpl) SysCreateOrganization(name string, requestType string, organizationDomains []string) (*model.Organization, error) {
	return s.app.sysCreateOrganization(name, requestType, organizationDomains)
}

func (s *systemImpl) SysUpdateOrganization(ID string, name string, requestType string, organizationDomains []string) error {
	return s.app.sysUpdateOrganization(ID, name, requestType, organizationDomains)
}

func (s *systemImpl) SysGetOrganizations() ([]model.Organization, error) {
	return s.app.sysGetOrganizations()
}

func (s *systemImpl) SysGetOrganization(ID string) (*model.Organization, error) {
	return s.app.sysGetOrganization(ID)
}

func (s *systemImpl) SysCreateApplication(name string, multiTenant bool, requiresOwnUsers bool, maxLoginSessionDuration *int, identifier string, nameInType string, versions []string) (*model.Application, error) {
	return s.app.sysCreateApplication(name, multiTenant, requiresOwnUsers, maxLoginSessionDuration, identifier, nameInType, versions)
}

func (s *systemImpl) SysGetApplication(ID string) (*model.Application, error) {
	return s.app.sysGetApplication(ID)
}

func (s *systemImpl) SysGetApplications() ([]model.Application, error) {
	return s.app.sysGetApplications()
}

func (s *systemImpl) SysCreatePermission(name string, serviceID string, assigners *[]string) (*model.Permission, error) {
	return s.app.sysCreatePermission(name, serviceID, assigners)
}

func (s *systemImpl) SysUpdatePermission(name string, serviceID *string, assigners *[]string) (*model.Permission, error) {
	return s.app.sysUpdatePermission(name, serviceID, assigners)
}

func (s *systemImpl) SysCreateAppOrgRole(name string, appOrgID string, description string, permissionNames []string) (*model.AppOrgRole, error) {
	return s.app.sysCreateAppOrgRole(name, appOrgID, description, permissionNames)
}

func (s *systemImpl) SysGetAppConfigs(appID string, appOrgID string, versionNumbers *model.VersionNumbers) ([]model.ApplicationConfig, error) {
	return s.app.sysGetAppConfigs(appID, appOrgID, versionNumbers)
}

func (s *systemImpl) SysGetAppConfig(id string) (*model.ApplicationConfig, error) {
	return s.app.sysGetAppConfig(id)
}

func (s *systemImpl) SysCreateAppConfig(version string, appTypeIdentifier string, appOrgID string, data map[string]interface{}, versionNumbers model.VersionNumbers) (*model.ApplicationConfig, error) {
	//get the app type
	applicationType, err := s.app.storage.FindApplicationTypeByIdentifier(appTypeIdentifier)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeApplicationType, logutils.StringArgs(appTypeIdentifier), err)

	}
	if applicationType == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeApplicationType, logutils.StringArgs(appTypeIdentifier))
	}
	if len(applicationType.Versions) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeApplicationTypeVersionList, logutils.StringArgs(appTypeIdentifier))
	}

	for _, supportedVersion := range applicationType.Versions {
		if versionNumbers == supportedVersion.VersionNumbers {
			return s.app.sysCreateAppConfig(version, *applicationType, appOrgID, data, versionNumbers)
		}
	}

	return nil, errors.ErrorData(logutils.StatusInvalid, model.TypeApplicationConfigsVersion, logutils.StringArgs(version+" for app_type_identifier: "+appTypeIdentifier))
}

func (s *systemImpl) SysUpdateAppConfig(id string, version string, appTypeIdentifier string, data map[string]interface{}, versionNumbers model.VersionNumbers) error {
	applicationType, err := s.app.storage.FindApplicationTypeByIdentifier(appTypeIdentifier)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionFind, model.TypeApplicationType, logutils.StringArgs(appTypeIdentifier), err)

	}
	if applicationType == nil {
		return errors.ErrorData(logutils.StatusMissing, model.TypeApplicationType, logutils.StringArgs(appTypeIdentifier))
	}
	if len(applicationType.Versions) == 0 {
		return errors.ErrorData(logutils.StatusMissing, model.TypeApplicationTypeVersionList, logutils.StringArgs(appTypeIdentifier))
	}

	for _, supportedVersion := range applicationType.Versions {
		if versionNumbers == supportedVersion.VersionNumbers {
			return s.app.sysUpdateAppConfig(id, version, data, versionNumbers)
		}
	}

	return errors.ErrorData(logutils.StatusInvalid, model.TypeApplicationConfigsVersion, logutils.StringArgs(appTypeIdentifier))
}

func (s *systemImpl) SysDeleteAppConfig(id string) error {
	return s.app.sysDeleteAppConfig(id)
}

func (s *systemImpl) SysGrantAccountPermissions(accountID string, permissionNames []string, assignerPermissions []string) error {
	return s.app.sysGrantAccountPermissions(accountID, permissionNames, assignerPermissions)
}

func (s *systemImpl) SysGrantAccountRoles(accountID string, appID string, roleIDs []string) error {
	return s.app.sysGrantAccountRoles(accountID, appID, roleIDs)
}

///
