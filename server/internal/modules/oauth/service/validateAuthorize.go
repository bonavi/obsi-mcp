package service

import (
	"server/internal/modules/oauth/model"
)

// ValidateAuthorize проверяет параметры запроса авторизации и возвращает клиента для формы
func (s *OAuthService) ValidateAuthorize(req model.AuthorizeReq) (model.Client, error) {

	// Клиент и адрес возврата: без них редиректить некуда
	client, found := s.parseClient(req.ClientID)
	if !found {
		return client, model.NewOAuthError("invalid_client", "unknown client_id")
	}
	registered := false
	for _, redirectURI := range client.RedirectURIs {
		registered = registered || sameRedirectURI(redirectURI, req.RedirectURI)
	}
	if !registered {
		return client, model.NewOAuthError("invalid_request", redirectNotRegistered)
	}

	// Остальные параметры
	if req.ResponseType != "code" {
		return client, model.NewOAuthError("unsupported_response_type", "only response_type=code is supported")
	}
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return client, model.NewOAuthError("invalid_request", "PKCE with code_challenge_method=S256 is required")
	}
	if req.Resource != "" && !sameResource(req.Resource, s.config.PublicURL) {
		return client, model.NewOAuthError("invalid_target", "resource does not match this server")
	}
	return client, nil
}
