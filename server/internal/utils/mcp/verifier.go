package mcp

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	sdkAuth "github.com/modelcontextprotocol/go-sdk/auth"
)

// NewTokenVerifier сравнивает Bearer-токен со статическим за постоянное время. Работает до появления OAuth
func NewTokenVerifier(token string) sdkAuth.TokenVerifier {
	wantedSum := sha256.Sum256([]byte(token))
	return func(_ context.Context, received string, _ *http.Request) (*sdkAuth.TokenInfo, error) {
		receivedSum := sha256.Sum256([]byte(received))
		if subtle.ConstantTimeCompare(receivedSum[:], wantedSum[:]) != 1 {
			return nil, sdkAuth.ErrInvalidToken
		}
		return &sdkAuth.TokenInfo{}, nil //nolint:exhaustruct
	}
}
