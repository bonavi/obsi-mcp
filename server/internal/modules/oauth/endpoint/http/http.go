// Package http — HTTP-эндпоинты Authorization Server: метаданные, регистрация, форма входа, токены
package http

import (
	"context"
	stdErrors "errors"
	"net/http"
	"pkg/log"
	"server/internal/modules/oauth/model"
	oauthService "server/internal/modules/oauth/service"
)

const maximalFormBytes = 64 << 10 // Максимальный размер тела запроса

var _ OAuthService = new(oauthService.OAuthService)

// OAuthService — интерфейс Authorization Server
type OAuthService interface {
	PublicURL() string
	RegisterClient(model.RegisterClientReq) (model.RegisterClientRes, error)
	ValidateAuthorize(model.AuthorizeReq) (model.Client, error)
	GrantCode(model.GrantCodeReq) (string, error)
	ExchangeCode(context.Context, model.ExchangeCodeReq) (model.TokenRes, error)
	Refresh(context.Context, string) (model.TokenRes, error)
}

// OAuthEndpoint — HTTP-обработчики OAuth
type OAuthEndpoint struct {
	oauthService OAuthService // Authorization Server
}

// NewOAuthEndpoint создает обработчики OAuth
func NewOAuthEndpoint(oauthService OAuthService) *OAuthEndpoint {
	return &OAuthEndpoint{oauthService: oauthService}
}

// Register добавляет маршруты OAuth в mux
func (e *OAuthEndpoint) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", e.Metadata)
	mux.HandleFunc("GET /.well-known/oauth-authorization-server/", e.Metadata)
	mux.HandleFunc("POST /oauth/register", e.RegisterClient)
	mux.HandleFunc("GET /oauth/authorize", e.AuthorizeForm)
	mux.HandleFunc("POST /oauth/authorize", e.AuthorizeSubmit)
	mux.HandleFunc("POST /oauth/token", e.Token)
}

// oauthError извлекает ошибку протокола; любая другая ошибка — server_error с записью в лог
func oauthError(err error) *model.OAuthError {
	var protocolError *model.OAuthError
	if stdErrors.As(err, &protocolError) {
		return protocolError
	}
	log.LogError(err)
	return model.NewOAuthError("server_error", "internal error")
}
