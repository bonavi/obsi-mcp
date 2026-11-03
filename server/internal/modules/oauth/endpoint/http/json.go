package http

import (
	"encoding/json"
	"net/http"
)

// writeJSON пишет JSON-ответ с нужным статусом
func writeJSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

// writeOAuthError пишет ошибку протокола (RFC 6749 §5.2)
func writeOAuthError(writer http.ResponseWriter, err error) {
	protocolError := oauthError(err)
	status := http.StatusBadRequest
	switch protocolError.Code {
	case "invalid_client":
		status = http.StatusUnauthorized
	case "server_error":
		status = http.StatusInternalServerError
	}
	writeJSON(writer, status, map[string]string{"error": protocolError.Code, "error_description": protocolError.Description})
}
