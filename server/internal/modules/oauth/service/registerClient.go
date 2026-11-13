package service

import (
	"server/internal/modules/oauth/model"
)

const (
	maximalNameLength  = 100          // Максимальная длина названия клиента
	maximalRedirectURI = 5            // Максимальное число адресов возврата
	defaultClientName  = "MCP-клиент" // Название, если клиент его не указал
)

// RegisterClient регистрирует клиента динамически. Состояние не создаётся: всё лежит в подписанном client_id
func (s *OAuthService) RegisterClient(req model.RegisterClientReq) (model.RegisterClientRes, error) {
	var res model.RegisterClientRes

	// Проверяем адреса возврата
	if len(req.RedirectURIs) == 0 || len(req.RedirectURIs) > maximalRedirectURI {
		return res, model.NewOAuthError("invalid_redirect_uri", "redirect_uris is required")
	}
	for _, redirectURI := range req.RedirectURIs {
		if !s.isAllowedRedirectURI(redirectURI) {
			return res, model.NewOAuthError("invalid_redirect_uri", "redirect_uri is not allowed")
		}
	}

	// Приводим название
	name := req.ClientName
	if name == "" {
		name = defaultClientName
	}
	if len([]rune(name)) > maximalNameLength {
		name = string([]rune(name)[:maximalNameLength])
	}

	// Выдаём подписанный идентификатор
	clientID, err := s.signClient(model.Client{Name: name, RedirectURIs: req.RedirectURIs})
	if err != nil {
		return res, err
	}
	return model.RegisterClientRes{
		ClientID:                clientID,
		ClientName:              name,
		RedirectURIs:            req.RedirectURIs,
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
	}, nil
}
