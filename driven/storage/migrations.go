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

package storage

import (
	"context"
	"core-building-block/core/model"
	"core-building-block/utils"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	migrationBatchSize int = 5000
)

// migrate to tenants accounts
func (sa *Adapter) migrateToTenantsAccounts() error {
	sa.logger.Debug("migrateToTenantsAccounts START")

	// only need phase 1 when there is only 1 account for a user in each org
	err := sa.startPhase1()
	if err != nil {
		return err
	}

	sa.logger.Debug("migrateToTenantsAccounts END")
	return nil
}

func (sa *Adapter) startPhase1() error {
	sa.logger.Debug("startPhase1 START")

	//all in transaction!
	transaction := func(context TransactionContext) error {

		//check if need to apply processing
		notMigratedCount, err := sa.findNotMigratedCount(context)
		if err != nil {
			return err
		}
		if *notMigratedCount == 0 {
			sa.logger.Debug("there is no what to be migrated, so do nothing")
			return nil
		}

		//WE MUST APPLY MIGRATION
		sa.logger.Debugf("there are %d accounts to be migrated", *notMigratedCount)

		//process duplicate events
		err = sa.processDuplicateAccounts(context)
		if err != nil {
			return err
		}

		return nil
	}

	err := sa.performTransactionWithTimeout(transaction, 10*time.Minute)
	if err != nil {
		return err
	}

	sa.logger.Debug("startPhase1 END")
	return nil
}

func (sa *Adapter) findNotMigratedCount(context TransactionContext) (*int64, error) {
	filter := bson.M{"migrated_2": bson.M{"$in": []interface{}{nil, false}}}
	count, err := sa.db.accounts.CountDocumentsWithContext(context, filter)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

func (sa *Adapter) processDuplicateAccounts(context TransactionContext) error {
	// batch the processing of the existing accounts
	batch := 0
	for {
		//find the duplicate accounts
		items, err := sa.findDuplicateAccounts(context, batch)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			sa.logger.Info("there is no more duplicated accounts")
			break
		}

		//construct tenants accounts
		tenantsAccounts, err := sa.constructTenantsAccounts(items)
		if err != nil {
			return err
		}

		//save tenants accounts
		err = sa.insertTenantAccounts(context, tenantsAccounts)
		if err != nil {
			return err
		}

		//mark the old accounts as processed
		accountsIDs := sa.getUniqueAccountsIDs(items)
		err = sa.markAccountsAsProcessed(context, accountsIDs)
		if err != nil {
			return err
		}

		batch++
	}

	return nil
}

func (sa *Adapter) getUniqueAccountsIDs(items map[string][]account) []string {
	var result []string

	for _, accounts := range items {
		for _, acc := range accounts {
			if !utils.Contains(result, acc.ID) {
				result = append(result, acc.ID)
			}
		}
	}

	return result
}

func (sa *Adapter) markAccountsAsProcessed(context TransactionContext, accountsIDs []string) error {
	filter := bson.D{primitive.E{Key: "_id", Value: bson.M{"$in": accountsIDs}}}

	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "migrated_2", Value: true},
		}},
	}

	_, err := sa.db.accounts.UpdateManyWithContext(context, filter, update, nil)
	if err != nil {
		return err
	}

	return nil
}

func (sa *Adapter) insertTenantAccounts(context TransactionContext, items []tenantAccount) error {

	stgItems := make([]interface{}, len(items))
	for i, p := range items {
		stgItems[i] = p
	}

	res, err := sa.db.tenantsAccounts.InsertManyWithContext(context, stgItems, nil)
	if err != nil {
		return err
	}

	if len(res.InsertedIDs) != len(items) {
		return errors.Newf("inserted:%d items:%d", len(res.InsertedIDs), len(items))
	}

	return nil
}

func (sa *Adapter) constructTenantsAccounts(duplicateAccounts map[string][]account) ([]tenantAccount, error) {
	//we need to load the apps orgs object from the database as we will need them
	allAppsOrgs, err := sa.getCachedApplicationOrganizations()
	if err != nil {
		return nil, err
	}

	//segment by org
	data := map[string][]orgAccounts{}
	for identifier, accounts := range duplicateAccounts {
		orgAccounts, err := sa.segmentByOrgID(allAppsOrgs, accounts)
		if err != nil {
			return nil, err
		}

		data[identifier] = orgAccounts
	}

	//use orgAccounts for easier manipulating
	orgIDsAccounts := sa.simplifyStructureData(data)

	res := []tenantAccount{}
	for _, item := range orgIDsAccounts {
		orgItems, err := sa.constructTenantsAccountsForOrg(item.OrgID, item.Accounts)
		if err != nil {
			return nil, err
		}
		res = append(res, orgItems...)
	}

	return res, nil
}

func (sa *Adapter) constructTenantsAccountsForOrg(orgID string, accounts []account) ([]tenantAccount, error) {
	//verify that there are no repeated identities across org applications
	notExist := sa.verifyNotExist(accounts)
	if !notExist {
		return nil, errors.Newf("%s has repeated items")
	}

	//process them
	resAccounts := []tenantAccount{}
	for _, account := range accounts {
		newTenantAccount := sa.createTenantAccount(orgID, account)
		resAccounts = append(resAccounts, newTenantAccount)
	}

	return resAccounts, nil
}

func (sa *Adapter) verifyNotExist(accounts []account) bool {
	for _, acc := range accounts {
		for _, acc2 := range accounts {
			if acc.ID == acc2.ID {
				continue //skip
			}

			if sa.containsIdentifier(acc.Identifiers, acc2.Identifiers) {
				// sa.logger.ErrorWithFields("duplicate identifier", logutils.Fields{"account1_id": acc.ID, "account2_id": acc2.ID})
				return false
			}
		}
	}
	return true
}

func (sa *Adapter) containsIdentifier(identifiers1 []accountIdentifier, identifiers2 []accountIdentifier) bool {
	for _, id := range identifiers1 {
		for _, id2 := range identifiers2 {
			if id.Identifier == id2.Identifier {
				// sa.logger.ErrorWithFields("duplicate identifier", logutils.Fields{"identifier1": id.Identifier, "identifier2": id2.Identifier})
				return true
			}
		}
	}
	return false
}

func (sa *Adapter) createTenantAccount(orgID string, account account) tenantAccount {

	id := account.ID
	scopes := account.Scopes
	identifiers := account.Identifiers
	authTypes := account.AuthTypes
	mfaTypes := account.MFATypes
	systemConfigs := account.SystemConfigs
	profile := account.Profile
	devices := account.Devices
	anonymous := account.Anonymous
	privacy := account.Privacy

	var verified *bool //not used?
	if account.Verified {
		verified = &account.Verified
	}

	dateCreated := account.DateCreated
	dateUpdated := account.DateUpdated
	isFollowing := account.IsFollowing
	lastLoginDate := account.LastLoginDate
	lastAccessTokenDate := account.LastAccessTokenDate

	//create org apps membership
	oaID := uuid.NewString()
	oaAppOrgID := account.AppOrgID
	oaPermissions := account.Permissions
	oaRoles := account.Roles
	oaGroups := account.Groups
	oaSecrets := account.Secrets
	oaPreferences := account.Preferences
	oaMostRecentClientVersion := account.MostRecentClientVersion

	orgAppsMemberships := []orgAppMembership{{ID: oaID, AppOrgID: oaAppOrgID,
		Permissions: oaPermissions, Roles: oaRoles, Groups: oaGroups, Secrets: oaSecrets,
		Preferences: oaPreferences, MostRecentClientVersion: oaMostRecentClientVersion}}

	return tenantAccount{ID: id, OrgID: orgID, OrgAppsMemberships: orgAppsMemberships, Scopes: scopes,
		Identifiers: identifiers, AuthTypes: authTypes, MFATypes: mfaTypes,
		SystemConfigs: systemConfigs, Profile: profile, Devices: devices, Anonymous: anonymous,
		Privacy: privacy, Verified: verified, DateCreated: dateCreated, DateUpdated: dateUpdated,
		IsFollowing: isFollowing, LastLoginDate: lastLoginDate, LastAccessTokenDate: lastAccessTokenDate}
}

func (sa *Adapter) simplifyStructureData(data map[string][]orgAccounts) []orgAccounts {
	temp := map[string][]account{}
	seen := []string{}
	for _, dataItem := range data {
		for _, orgAccounts := range dataItem {
			orgID := orgAccounts.OrgID
			orgAllAccounts := temp[orgID]

			for _, acc := range orgAccounts.Accounts {
				if !utils.Contains(seen, acc.ID) { // Check if already added
					seen = append(seen, acc.ID)
					orgAllAccounts = append(orgAllAccounts, acc)
				}
			}

			temp[orgID] = orgAllAccounts
		}
	}

	//prepare response
	res := []orgAccounts{}
	for orgID, tempItem := range temp {
		res = append(res, orgAccounts{OrgID: orgID, Accounts: tempItem})
	}
	return res
}

type orgAccounts struct {
	OrgID    string
	Accounts []account
}

func (sa *Adapter) segmentByOrgID(allAppsOrgs []model.ApplicationOrganization, accounts []account) ([]orgAccounts, error) {
	tempMap := map[string][]account{}
	for _, account := range accounts {
		currentOrgID, err := sa.findOrgIDByAppOrgID(account.AppOrgID, allAppsOrgs)
		if err != nil {
			return nil, err
		}

		orgAccountsMap := tempMap[currentOrgID]
		orgAccountsMap = append(orgAccountsMap, account)
		tempMap[currentOrgID] = orgAccountsMap

	}

	result := []orgAccounts{}
	for orgID, accounts := range tempMap {
		current := orgAccounts{OrgID: orgID, Accounts: accounts}
		result = append(result, current)
	}

	return result, nil
}

func (sa *Adapter) findOrgIDByAppOrgID(appOrgID string, allAppsOrgs []model.ApplicationOrganization) (string, error) {
	for _, item := range allAppsOrgs {
		if item.ID == appOrgID {
			return item.Organization.ID, nil
		}
	}
	return "", errors.Newf("no org for app org id - %s", appOrgID)
}

func (sa *Adapter) findDuplicateAccounts(context TransactionContext, batch int) (map[string][]account, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{"migrated_2": bson.M{"$in": []interface{}{nil, false}}}, //iterate only not migrated records
		},
		// batch the matched accounts according to migrationBatchSize
		{
			"$skip": batch * migrationBatchSize,
		},
		{
			"$limit": migrationBatchSize,
		},
		{
			"$unwind": "$identifiers",
		},
		{
			"$group": bson.M{
				"_id": "$identifiers.identifier",
				"accounts": bson.M{
					"$push": bson.M{
						"id": "$_id",
					},
				},
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"result": bson.M{
					"$push": bson.M{
						"k": "$_id",
						"v": bson.M{
							"accounts": "$accounts",
						},
					},
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$arrayToObject": "$result",
				},
			},
		},
	}

	cursor, err := sa.db.accounts.coll.Aggregate(context, pipeline)
	if err != nil {
		return nil, err
	}

	var result bson.M
	if cursor.Next(context) {
		err := cursor.Decode(&result)
		if err != nil {
			return nil, err
		}
	}

	if len(result) == 0 {

		return nil, nil
	}

	var resTypeResult []identityAccountsItem

	for key, value := range result {
		valueM := value.(primitive.M)
		accountsArr := valueM["accounts"].(primitive.A)

		var accounts []accountItem

		for _, element := range accountsArr {
			accountObj := element.(primitive.M)

			var account accountItem
			account.ID = accountObj["id"].(string)

			accounts = append(accounts, account)
		}

		item := identityAccountsItem{
			Identifier: key,
			Accounts:   accounts,
		}

		resTypeResult = append(resTypeResult, item)
	}

	//prepare founded duplicate accounts
	preparedResponse, err := sa.prepareFoundedDuplicateAccounts(context, resTypeResult)
	if err != nil {
		return nil, err
	}

	return preparedResponse, nil
}

type accountItem struct {
	ID string `bson:"id"`
}
type identityAccountsItem struct {
	Identifier string        `bson:"id"`
	Accounts   []accountItem `bson:"accounts"`
}

func (sa *Adapter) prepareFoundedDuplicateAccounts(context TransactionContext, foundedItems []identityAccountsItem) (map[string][]account, error) {

	if len(foundedItems) == 0 {
		return nil, nil
	}

	//load all accounts
	accountsIDs := []string{}
	for _, item := range foundedItems {
		accounts := item.Accounts
		for _, acc := range accounts {
			accountsIDs = append(accountsIDs, acc.ID)
		}
	}
	findFilter := bson.M{"_id": bson.M{"$in": accountsIDs}}
	var accounts []account
	err := sa.db.accounts.FindWithContext(context, findFilter, &accounts, nil)
	if err != nil {
		return nil, err
	}

	//prepare result
	result := map[string][]account{}
	for _, item := range foundedItems {
		identifier := item.Identifier
		accountsIDs := item.Accounts

		resAccounts, err := sa.getFullAccountsObjects(accountsIDs, accounts)
		if err != nil {
			return nil, err
		}
		result[identifier] = resAccounts
	}

	return result, nil
}

func (sa *Adapter) getFullAccountsObjects(accountsIDs []accountItem, allAccounts []account) ([]account, error) {
	result := []account{}
	for _, item := range accountsIDs {
		//find the full account object
		var resAccount *account
		for _, acc := range allAccounts {
			if item.ID == acc.ID {
				resAccount = &acc
				break
			}
		}

		if resAccount == nil {
			return nil, errors.Newf("cannot find full account for %s", item.ID)
		}
		result = append(result, *resAccount)
	}

	return result, nil
}

func (sa *Adapter) performTransactionWithTimeout(transaction func(context TransactionContext) error, timeout time.Duration) error {
	// Setting a timeout for the transaction
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// transaction
	err := sa.db.dbClient.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionStart, logutils.TypeTransaction, nil, err)
		}

		err = transaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
		}

		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			sa.abortTransaction(sessionContext)
			return errors.WrapErrorAction(logutils.ActionCommit, logutils.TypeTransaction, nil, err)
		}
		return nil
	})

	return err
}
