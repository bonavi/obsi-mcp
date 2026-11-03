package config

import (
	"pkg/env"
	"pkg/log"
	"time"
)

// Config — переменные окружения. Все обязательные: без значения сервис не стартует
type Config struct {
	ServiceName         string `env:"SERVICE_NAME,notEmpty"`  // Имя сервиса в логах
	Environment         string `env:"ENV,notEmpty"`           // Окружение
	IsStackTraceEnabled bool   `env:"IS_STACK_TRACE_ENABLED"` // Признак записи стектрейсов
	AuthToken           string `env:"AUTH_TOKEN,notEmpty"`    // Статический токен доступа до появления OAuth

	Listen struct {
		ServiceHTTP string `env:"LISTEN_SERVICE_HTTP,notEmpty"` // Адрес служебного HTTP-сервера
		MCP         string `env:"LISTEN_MCP,notEmpty"`          // Адрес MCP-сервера
	}

	Logger log.LoggerSettingsEnv // Настройки логгера

	Vault struct {
		Directory string   `env:"VAULT_DIR,notEmpty"`        // Папка с клоном базы
		DenyPaths []string `env:"VAULT_DENY_PATHS,notEmpty"` // Папки, недоступные через инструменты
	}

	Git struct {
		Remote          string        `env:"GIT_REMOTE,notEmpty"`       // Имя удалённого репозитория
		Branch          string        `env:"GIT_BRANCH,notEmpty"`       // Рабочая ветка
		AuthorName      string        `env:"GIT_AUTHOR_NAME,notEmpty"`  // Имя автора коммитов
		AuthorEmail     string        `env:"GIT_AUTHOR_EMAIL,notEmpty"` // Почта автора коммитов
		SyncInterval    time.Duration `env:"SYNC_INTERVAL,notEmpty"`    // Период синхронизации с удалённым репозиторием
		AutoCommitAfter time.Duration `env:"AUTOCOMMIT_AFTER,notEmpty"` // Возраст правки, после которого сервер коммитит её сам
	}
}

// Load загружает конфиг из переменных окружения
func Load() Config {
	return env.Load[Config]()
}
