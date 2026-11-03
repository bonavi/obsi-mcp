package mcp

import (
	"net/http"
	"pkg/log"

	sdkAuth "github.com/modelcontextprotocol/go-sdk/auth"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

const (
	maxRequestBytes      = 2 << 20                                 // Максимальный размер запроса
	ResourceMetadataPath = "/.well-known/oauth-protected-resource" // Путь метаданных защищенного ресурса (RFC 9728)
)

// NewHandler оборачивает MCP-сервер в HTTP: Bearer-токен проверяется до транспорта, затем
// Streamable HTTP без сессий — каждый запрос независим и переживает перезапуск пода
func NewHandler(server *sdk.Server, sessionChecker SessionChecker, publicURL string) http.Handler {
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

	requireToken := sdkAuth.RequireBearerToken(NewTokenVerifier(sessionChecker), &sdkAuth.RequireBearerTokenOptions{ //nolint:exhaustruct
		ResourceMetadataURL:    publicURL + ResourceMetadataPath,
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

// NewResourceMetadataHandler отдает метаданные защищенного ресурса: клиент узнает из них, где
// Authorization Server. У нас это тот же хост
func NewResourceMetadataHandler(publicURL string) http.Handler {
	return sdkAuth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{ //nolint:exhaustruct
		Resource:               publicURL,
		AuthorizationServers:   []string{publicURL},
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Obsidian",
	})
}
