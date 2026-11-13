// Package service — Authorization Server OAuth 2.1 для одного пользователя без базы данных
package service

import (
	"server/internal/modules/oauth/model"
	oauthRepository "server/internal/modules/oauth/repository"
	"strings"
	"sync"
	"time"
)

var _ OAuthRepository = new(oauthRepository.OAuthRepository)

// OAuthRepository — хранение кодов и сессий
type OAuthRepository interface {
	CreateCode(model.Code, time.Time)
	ConsumeCode(string, time.Time) (model.Code, bool)
	CreateSession(string, model.Session, time.Time) error
	RotateSession(string, string, time.Time, time.Time) (string, bool, error)
	IsFamilyActive(string, time.Time) bool
}

// Config — параметры Authorization Server
type Config struct {
	PublicURL       string        // Канонический URL MCP, он же issuer и resource
	AccessTokenTTL  time.Duration // Срок жизни access-токена
	RefreshTokenTTL time.Duration // Срок жизни refresh-токена
	PasswordHash    string        // Хэш пароля в формате argon2id
	ClientKey       []byte        // Ключ подписи client_id

	LoopbackRedirectPaths []string // Разрешённые пути возврата локальных клиентов на localhost
}

// OAuthService — Authorization Server
type OAuthService struct {
	oauthRepository OAuthRepository // Коды и сессии
	config          Config          // Параметры
	password        passwordParams  // Разобранный хэш пароля
	passwordLock    sync.Mutex      // Проверки идут по одной: argon2 тяжёлый для Pi
	failedAttempts  int             // Неудачных попыток подряд
	lockedUntil     time.Time       // До какого момента вход закрыт
}

// NewOAuthService создает Authorization Server
func NewOAuthService(oauthRepository OAuthRepository, config Config) (*OAuthService, error) {
	password, err := parsePasswordHash(config.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &OAuthService{
		oauthRepository: oauthRepository,
		config:          config,
		password:        password,
		passwordLock:    sync.Mutex{},
		failedAttempts:  0,
		lockedUntil:     time.Time{},
	}, nil
}

// PublicURL возвращает канонический URL MCP
func (s *OAuthService) PublicURL() string {
	return s.config.PublicURL
}

// IsSessionActive проверяет, не отозван ли вход
func (s *OAuthService) IsSessionActive(familyID string) bool {
	return s.oauthRepository.IsFamilyActive(familyID, time.Now())
}

// sameResource сравнивает адреса ресурса без завершающего слэша
func sameResource(first, second string) bool {
	return strings.TrimSuffix(first, "/") == strings.TrimSuffix(second, "/")
}
