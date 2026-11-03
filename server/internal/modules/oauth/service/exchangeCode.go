package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"server/internal/modules/oauth/model"
	"time"
)

// ExchangeCode меняет код на токены: проверяет клиента, адрес возврата, PKCE и resource
func (s *OAuthService) ExchangeCode(ctx context.Context, req model.ExchangeCodeReq) (model.TokenRes, error) {
	now := time.Now()

	// Код одноразовый: любая неудачная попытка его тоже сжигает
	code, found := s.oauthRepository.ConsumeCode(req.Code, now)
	if !found {
		return model.TokenRes{}, model.NewOAuthError("invalid_grant", "code is unknown, expired or already used")
	}
	if subtle.ConstantTimeCompare([]byte(req.ClientID), []byte(code.ClientID)) != 1 {
		return model.TokenRes{}, model.NewOAuthError("invalid_grant", "client_id does not match the code")
	}
	if req.RedirectURI != code.RedirectURI {
		return model.TokenRes{}, model.NewOAuthError("invalid_grant", "redirect_uri does not match the code")
	}
	if req.Resource != "" && !sameResource(req.Resource, code.Resource) {
		return model.TokenRes{}, model.NewOAuthError("invalid_target", "resource does not match the code")
	}

	// PKCE: SHA256(verifier) должен совпасть с challenge
	sum := sha256.Sum256([]byte(req.CodeVerifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if req.CodeVerifier == "" || subtle.ConstantTimeCompare([]byte(challenge), []byte(code.CodeChallenge)) != 1 {
		return model.TokenRes{}, model.NewOAuthError("invalid_grant", "code_verifier does not match code_challenge")
	}

	// Начинаем новую цепочку токенов
	familyID, err := randomToken()
	if err != nil {
		return model.TokenRes{}, err
	}
	refreshToken, err := randomToken()
	if err != nil {
		return model.TokenRes{}, err
	}
	session := model.Session{FamilyID: familyID, IsUsed: false, ExpiresAt: now.Add(s.config.RefreshTokenTTL)}
	if err = s.oauthRepository.CreateSession(hashToken(refreshToken), session, now); err != nil {
		return model.TokenRes{}, err
	}
	return s.issueTokens(ctx, familyID, refreshToken)
}
