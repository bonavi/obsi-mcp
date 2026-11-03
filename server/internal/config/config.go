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

	Listen struct {
		ServiceHTTP string `env:"LISTEN_SERVICE_HTTP,notEmpty"` // Адрес служебного HTTP-сервера
		MCP         string `env:"LISTEN_MCP,notEmpty"`          // Адрес MCP-сервера
	}

	// MCP-слой: авторизация по OAuth 2.1
	MCP struct {
		PublicURL string `env:"MCP_PUBLIC_URL,notEmpty"` // Канонический публичный URL без завершающего слэша
	}

	// Вход единственного пользователя и токены
	Auth struct {
		PasswordHash    string        `env:"AUTH_PASSWORD_HASH,notEmpty"`     // Хэш пароля argon2id, см. cmd/hashPassword
		SigningKey      string        `env:"AUTH_TOKEN_SIGNING_KEY,notEmpty"` // Ключ подписи токенов
		AccessTokenTTL  time.Duration `env:"AUTH_ACCESS_TOKEN_TTL,notEmpty"`  // Срок жизни access-токена
		RefreshTokenTTL time.Duration `env:"AUTH_REFRESH_TOKEN_TTL,notEmpty"` // Срок жизни refresh-токена
		StateFile       string        `env:"AUTH_STATE_FILE,notEmpty"`        // Файл с refresh-сессиями, вне базы заметок
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
