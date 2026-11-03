package mcp

import (
	"net/http"
	"pkg/log"

	sdkAuth "github.com/modelcontextprotocol/go-sdk/auth"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxRequestBytes = 2 << 20 // Максимальный размер запроса

// NewHandler оборачивает MCP-сервер в HTTP: Bearer-токен проверяется до транспорта, затем
// Streamable HTTP без сессий — каждый запрос независим и переживает перезапуск пода
func NewHandler(server *sdk.Server, token string) http.Handler {
	transport := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{ //nolint:exhaustruct
			Stateless:    true,
			JSONResponse: true,
		},
	)

	// Ограничиваем размер запроса
	limited := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBytes)
		transport.ServeHTTP(writer, request)
	})

	requireToken := sdkAuth.RequireBearerToken(NewTokenVerifier(token), &sdkAuth.RequireBearerTokenOptions{ //nolint:exhaustruct
		AllowMissingExpiration: true,
	})
	protected := requireToken(limited)

	// Отказы пишем в журнал без токена
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		protected.ServeHTTP(recorder, request)
		if recorder.status == http.StatusUnauthorized {
			log.Warning("MCP: отклонён запрос без действующего токена")
		}
	})
}
