package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"server/internal/utils/errors"
)

const randomTokenBytes = 32 // Размер случайных токенов

// randomToken создаёт случайную строку
func randomToken() (string, error) {
	raw := make([]byte, randomTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", errors.InternalServer.Wrap(err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// hashToken считает хэш refresh-токена для хранения
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
