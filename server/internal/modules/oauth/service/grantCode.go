package service

import (
	"server/internal/modules/oauth/model"
	"time"
)

const codeTTL = 10 * time.Minute // Срок жизни кода авторизации

// GrantCode проверяет пароль и выдаёт одноразовый код
func (s *OAuthService) GrantCode(req model.GrantCodeReq) (string, error) {

	// Параметры проверяем повторно: форма могла быть подделана
	if _, err := s.ValidateAuthorize(req.Authorize); err != nil {
		return "", err
	}
	now := time.Now()
	if err := s.checkPassword(req.Password, now); err != nil {
		return "", err
	}

	// Выдаём код
	value, err := randomToken()
	if err != nil {
		return "", err
	}
	s.oauthRepository.CreateCode(model.Code{
		Value:         value,
		ClientID:      req.Authorize.ClientID,
		RedirectURI:   req.Authorize.RedirectURI,
		CodeChallenge: req.Authorize.CodeChallenge,
		Resource:      s.config.PublicURL,
		ExpiresAt:     now.Add(codeTTL),
	}, now)
	return value, nil
}
