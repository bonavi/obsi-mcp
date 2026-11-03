package http

import (
	"encoding/json"
	"net/http"
	"server/internal/modules/oauth/model"
)

// RegisterClient регистрирует клиента динамически (RFC 7591)
func (e *OAuthEndpoint) RegisterClient(writer http.ResponseWriter, request *http.Request) {
	var req model.RegisterClientReq
	if err := json.NewDecoder(http.MaxBytesReader(writer, request.Body, maximalFormBytes)).Decode(&req); err != nil {
		writeOAuthError(writer, model.NewOAuthError("invalid_client_metadata", "body must be JSON"))
		return
	}
	res, err := e.oauthService.RegisterClient(req)
	if err != nil {
		writeOAuthError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, res)
}
