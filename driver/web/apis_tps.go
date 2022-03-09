package web

import (
	"core-building-block/core"
	"core-building-block/core/model"
	Def "core-building-block/driver/web/docs/gen"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/tokenauth"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
)

//TPSApisHandler handles the APIs implementation used by third-party services
type TPSApisHandler struct {
	coreAPIs *core.APIs
}

func (h TPSApisHandler) getServiceRegistrations(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HttpResponse {
	serviceIDsParam := r.URL.Query().Get("ids")
	if serviceIDsParam == "" {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("ids"), nil, http.StatusBadRequest, false)
	}
	serviceIDs := strings.Split(serviceIDsParam, ",")

	serviceRegs, err := h.coreAPIs.Auth.GetServiceRegistrations(serviceIDs)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, model.TypeServiceReg, nil, err, http.StatusInternalServerError, true)
	}

	serviceRegResp := authServiceRegListToDef(serviceRegs)

	data, err := json.Marshal(serviceRegResp)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, model.TypeServiceReg, nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

func (h TPSApisHandler) getAuthKeys(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HttpResponse {
	keys, err := h.coreAPIs.Auth.GetAuthKeySet()
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, model.TypeJSONWebKeySet, logutils.StringArgs("auth"), err, http.StatusInternalServerError, true)
	}

	keysResp := jsonWebKeySetDef(keys)

	data, err := json.Marshal(keysResp)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, model.TypeJSONWebKeySet, nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

func (h TPSApisHandler) getServiceAccountParams(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HttpResponse {
	params := mux.Vars(r)
	accountID := params["id"]
	if len(accountID) <= 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	message, accountParams, err := h.coreAPIs.Auth.GetServiceAccountParams(accountID, r, l)
	if err != nil {
		if message != nil {
			return l.HttpResponseError(*message, err, http.StatusUnauthorized, false)
		}
		return l.HttpResponseError("Error getting service account params", err, http.StatusInternalServerError, false)
	}

	appOrgPairs := appOrgPairListToDef(accountParams)

	respData, err := json.Marshal(appOrgPairs)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, logutils.MessageDataType("service account params response"), nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(respData)
}

func (h TPSApisHandler) getServiceAccessToken(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HttpResponse {
	message, accessToken, err := h.coreAPIs.Auth.GetServiceAccessToken(r, l)
	if err != nil {
		if message != nil {
			return l.HttpResponseError(*message, err, http.StatusUnauthorized, false)
		}
		return l.HttpResponseError("Error getting access token", err, http.StatusInternalServerError, false)
	}

	tokenType := Def.SharedResRokwireTokenTokenTypeBearer
	rokwireToken := Def.SharedResRokwireToken{AccessToken: &accessToken, TokenType: &tokenType}

	respData, err := json.Marshal(rokwireToken)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, logutils.MessageDataType("service access token response"), nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(respData)
}

//NewTPSApisHandler creates new tps Handler instance
func NewTPSApisHandler(coreAPIs *core.APIs) TPSApisHandler {
	return TPSApisHandler{coreAPIs: coreAPIs}
}
