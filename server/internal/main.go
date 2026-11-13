package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"pkg/gracefulShutdown"
	fiber2 "pkg/http/fiber"
	"pkg/jwtManager"
	"pkg/log"
	"pkg/log/model"
	"pkg/middleware"
	"pkg/panicRecover"
	"pkg/stackTrace"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"golang.org/x/sync/errgroup"

	"server/internal/config"
	gitSyncEndpointMCP "server/internal/modules/gitSync/endpoint/mcp"
	gitSyncModel "server/internal/modules/gitSync/model"
	gitSyncRepository "server/internal/modules/gitSync/repository"
	gitSyncService "server/internal/modules/gitSync/service"
	noteEndpointMCP "server/internal/modules/note/endpoint/mcp"
	noteRepository "server/internal/modules/note/repository"
	noteService "server/internal/modules/note/service"
	oauthEndpointHTTP "server/internal/modules/oauth/endpoint/http"
	oauthRepository "server/internal/modules/oauth/repository"
	oauthService "server/internal/modules/oauth/service"
	"server/internal/utils/errors"
	mcpTransport "server/internal/utils/mcp"
)

const minimalSigningKeyLength = 32 // Минимальная длина ключа подписи

const version = "@{version}"
const build = "@{build}"
const buildDate = "@{buildDate}"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run собирает зависимости и запускает сервер
func run() error {

	// Перехватываем возможную панику
	defer panicRecover.PanicRecover(func(err error) {
		log.Fatal(err)
	})

	// Основной контекст приложения
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Загружаем конфиг и настраиваем стектрейсы
	conf := config.Load()
	stackTrace.Init("", conf.IsStackTraceEnabled, 1)

	// Получаем имя тачки
	hostname, err := os.Hostname()
	if err != nil {
		return errors.InternalServer.Wrap(err)
	}

	// Инициализируем логгер
	if err = log.InitDefaultLogger(model.SystemInfo{
		Version:     version,
		Build:       build,
		ServiceName: conf.ServiceName,
		Env:         conf.Environment,
		Hostname:    hostname,
		BuildDate:   buildDate,
	}, conf.Logger); err != nil {
		return err
	}

	// Собираем модуль синхронизации
	gitSync := gitSyncService.NewGitSyncService(
		gitSyncRepository.NewGitSyncRepository(gitSyncModel.Config{
			Directory:   conf.Vault.Directory,
			Remote:      conf.Git.Remote,
			Branch:      conf.Git.Branch,
			AuthorName:  conf.Git.AuthorName,
			AuthorEmail: conf.Git.AuthorEmail,
		}),
		conf.Git.SyncInterval,
		conf.Git.AutoCommitAfter,
	)

	// Собираем модуль заметок
	notes, err := noteRepository.NewNoteRepository(conf.Vault.Directory, conf.Vault.DenyPaths)
	if err != nil {
		return err
	}

	// Настраиваем OAuth
	if err = validateAuthConfig(conf); err != nil {
		return err
	}
	jwtManager.Init([]byte(conf.Auth.SigningKey), conf.Auth.AccessTokenTTL, conf.Auth.RefreshTokenTTL)
	sessions, err := oauthRepository.NewOAuthRepository(conf.Auth.StateFile)
	if err != nil {
		return err
	}
	oauth, err := oauthService.NewOAuthService(sessions, oauthService.Config{
		PublicURL:       conf.MCP.PublicURL,
		AccessTokenTTL:  conf.Auth.AccessTokenTTL,
		RefreshTokenTTL: conf.Auth.RefreshTokenTTL,
		PasswordHash:    conf.Auth.PasswordHash,
		ClientKey:       deriveClientKey(conf.Auth.SigningKey),

		LoopbackRedirectPaths: conf.Auth.LoopbackRedirectPaths,
	})
	if err != nil {
		return err
	}

	// Настраиваем MCP-сервер: инструменты регистрируют модули в своих endpoint/mcp
	mcpServer := mcpTransport.NewServer(version)
	noteEndpointMCP.Register(mcpServer, noteService.NewNoteService(notes, gitSync))
	gitSyncEndpointMCP.Register(mcpServer, gitSync)
	mcpMux := http.NewServeMux()
	oauthEndpointHTTP.NewOAuthEndpoint(oauth).Register(mcpMux)
	mcpMux.Handle("GET "+mcpTransport.ResourceMetadataPath, mcpTransport.NewResourceMetadataHandler(conf.MCP.PublicURL))
	mcpMux.Handle("GET "+mcpTransport.ResourceMetadataPath+"/", mcpTransport.NewResourceMetadataHandler(conf.MCP.PublicURL))
	mcpMux.Handle("/", mcpTransport.NewHandler(mcpServer, oauth, conf.MCP.PublicURL))
	mcpHTTPServer, err := fiber2.GetDefaultServer(conf.ServiceName, nil)
	if err != nil {
		return err
	}
	mcpHTTPServer.All("/*", adaptor.HTTPHandler(mcpMux))

	// Настраиваем служебный HTTP-сервер
	serviceHTTPServer, err := fiber2.GetDefaultServer(conf.ServiceName, nil)
	if err != nil {
		return err
	}
	serviceHTTPServer.Use(healthcheck.New(healthcheck.Config{
		Next:              nil,
		LivenessProbe:     nil,
		LivenessEndpoint:  "",
		ReadinessProbe:    fiber2.NewReadyHandler(nil),
		ReadinessEndpoint: "",
	}))
	serviceHTTPServer.Get("/version", middleware.NewVersionHandler(version, build, buildDate, hostname))

	// Создаем wait группу
	eg, ctx := errgroup.WithContext(ctx)

	// Создаем горутину фоновой синхронизации
	eg.Go(func() error {
		gitSync.Run(ctx)
		return nil
	})

	// Создаем горутину на запуск MCP-сервера
	eg.Go(func() error {
		log.Info(fmt.Sprintf("MCP server is listening :%v", conf.Listen.MCP))
		if err := mcpHTTPServer.Listen(conf.Listen.MCP); err != nil {
			return errors.InternalServer.Wrap(err)
		}
		return nil
	})

	// Создаем горутину на запуск служебного HTTP-сервера
	eg.Go(func() error {
		log.Info(fmt.Sprintf("service HTTP server is listening :%v", conf.Listen.ServiceHTTP))
		if err := serviceHTTPServer.Listen(conf.Listen.ServiceHTTP); err != nil {
			return errors.InternalServer.Wrap(err)
		}
		return nil
	})

	// Добавляем в группу горутину для graceful shutdown
	gracefulShutdown.AddGracefulShutdownErrGroup(eg, ctx, []*fiber.App{mcpHTTPServer, serviceHTTPServer}, nil)

	// Ждем завершения контекста или ошибок в горутинах
	return eg.Wait()
}

// validateAuthConfig не даёт стартовать со слабым ключом или файлом сессий внутри базы:
// git add -A отправил бы его на GitHub
func validateAuthConfig(conf config.Config) error {
	if len(conf.Auth.SigningKey) < minimalSigningKeyLength {
		return errors.InternalServer.New("AUTH_TOKEN_SIGNING_KEY is too short")
	}
	relativePath, err := filepath.Rel(conf.Vault.Directory, conf.Auth.StateFile)
	if err != nil {
		return errors.InternalServer.Wrap(err)
	}
	if relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return errors.InternalServer.New("AUTH_STATE_FILE must be outside VAULT_DIR")
	}
	return nil
}

// deriveClientKey выводит отдельный ключ подписи client_id, чтобы он не совпадал с ключом токенов
func deriveClientKey(signingKey string) []byte {
	signer := hmac.New(sha256.New, []byte(signingKey))
	signer.Write([]byte("oauth-client-id"))
	return signer.Sum(nil)
}
