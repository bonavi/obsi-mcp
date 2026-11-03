package service

import (
	"context"
	"server/internal/modules/oauth/model"
	"time"
)

// Refresh одноразово меняет refresh-токен на новую пару
func (s *OAuthService) Refresh(ctx context.Context, refreshToken string) (model.TokenRes, error) {
	now := time.Now()
	newToken, err := randomToken()
	if err != nil {
		return model.TokenRes{}, err
	}
	familyID, allowed, err := s.oauthRepository.RotateSession(hashToken(refreshToken), hashToken(newToken), now.Add(s.config.RefreshTokenTTL), now)
	if err != nil {
		return model.TokenRes{}, err
	}
	if !allowed {
		return model.TokenRes{}, model.NewOAuthError("invalid_grant", "refresh token is invalid or revoked")
	}
	return s.issueTokens(ctx, familyID, newToken)
}
