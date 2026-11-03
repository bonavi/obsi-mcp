package mcp

import (
	"context"
	"fmt"
	"net/http"
	"pkg/jwtManager"
	"server/internal/utils/auth"

	sdkAuth "github.com/modelcontextprotocol/go-sdk/auth"
)

// SessionChecker проверяет, не отозван ли вход
type SessionChecker interface {
	IsSessionActive(string) bool
}

// NewTokenVerifier проверяет Bearer-токен как access-JWT и что его цепочка не отозвана.
// Подпись и срок проверяет jwtManager, поэтому Expiration не заполняется (AllowMissingExpiration)
func NewTokenVerifier(sessionChecker SessionChecker) sdkAuth.TokenVerifier {
	return func(ctx context.Context, token string, _ *http.Request) (*sdkAuth.TokenInfo, error) {
		claims, err := jwtManager.ParseToken[auth.Claims](ctx, token, jwtManager.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid token", sdkAuth.ErrInvalidToken)
		}
		if claims.FamilyID == "" || !sessionChecker.IsSessionActive(claims.FamilyID) {
			return nil, fmt.Errorf("%w: session revoked", sdkAuth.ErrInvalidToken)
		}
		return &sdkAuth.TokenInfo{}, nil //nolint:exhaustruct
	}
}
