package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"server/internal/modules/oauth/model"
	"server/internal/utils/errors"
	"strings"
)

// signClient упаковывает данные клиента в подписанный client_id: хранить регистрации не нужно
func (s *OAuthService) signClient(client model.Client) (string, error) {
	payload, err := json.Marshal(client)
	if err != nil {
		return "", errors.InternalServer.Wrap(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + s.clientSignature(encoded), nil
}

// parseClient проверяет подпись client_id и возвращает данные клиента
func (s *OAuthService) parseClient(clientID string) (model.Client, bool) {
	var client model.Client
	encoded, signature, found := strings.Cut(clientID, ".")
	if !found || !hmac.Equal([]byte(signature), []byte(s.clientSignature(encoded))) {
		return client, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || json.Unmarshal(payload, &client) != nil {
		return client, false
	}
	return client, true
}

// clientSignature считает подпись данных клиента
func (s *OAuthService) clientSignature(encoded string) string {
	signer := hmac.New(sha256.New, s.config.ClientKey)
	signer.Write([]byte(encoded))
	return base64.RawURLEncoding.EncodeToString(signer.Sum(nil))
}
