package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"pkg/gracefulShutdown"
	fiber2 "pkg/http/fiber"
	"pkg/log"
	"pkg/log/model"
	"pkg/middleware"
	"pkg/panicRecover"
	"pkg/stackTrace"
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
	"server/internal/utils/errors"
	mcpTransport "server/internal/utils/mcp"
)

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

	// Настраиваем MCP-сервер: инструменты регистрируют модули в своих endpoint/mcp
	mcpServer := mcpTransport.NewServer(version)
	noteEndpointMCP.Register(mcpServer, noteService.NewNoteService(notes, gitSync))
	gitSyncEndpointMCP.Register(mcpServer, gitSync)
	mcpHTTPServer, err := fiber2.GetDefaultServer(conf.ServiceName, nil)
	if err != nil {
		return err
	}
	mcpHTTPServer.All("/*", adaptor.HTTPHandler(mcpTransport.NewHandler(mcpServer, conf.AuthToken)))

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
