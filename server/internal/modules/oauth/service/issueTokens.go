package service

import (
	"context"
	"pkg/jwtManager"
	"server/internal/modules/oauth/model"
	"server/internal/utils/auth"
	"server/internal/utils/errors"
)

// issueTokens собирает ответ с access-JWT и refresh-токеном
func (s *OAuthService) issueTokens(ctx context.Context, familyID, refreshToken string) (model.TokenRes, error) {
	accessToken, err := jwtManager.GenerateToken(ctx, jwtManager.AccessToken, auth.Claims{FamilyID: familyID})
	if err != nil {
		return model.TokenRes{}, errors.InternalServer.Wrap(err)
	}
	return model.TokenRes{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
		RefreshToken: refreshToken,
	}, nil
}
