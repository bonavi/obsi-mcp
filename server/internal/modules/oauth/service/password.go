package service

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"server/internal/modules/oauth/model"
	"server/internal/utils/errors"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	freeAttempts   = 5                // Попыток без блокировки
	firstLockout   = time.Minute      // Первая блокировка
	maximalLockout = 15 * time.Minute // Самая долгая блокировка
	minimalMemory  = 32 * 1024        // Минимальная память argon2 в КиБ
)

// passwordParams — разобранный хэш argon2id
type passwordParams struct {
	memory      uint32 // Память в КиБ
	iterations  uint32 // Число проходов
	parallelism uint8  // Число потоков
	salt        []byte // Соль
	hash        []byte // Хэш
}

// parsePasswordHash разбирает строку $argon2id$v=19$m=...,t=...,p=...$соль$хэш
func parsePasswordHash(encoded string) (passwordParams, error) {
	var params passwordParams
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return params, errors.InternalServer.New("password hash must be argon2id")
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.iterations, &params.parallelism); err != nil {
		return params, errors.InternalServer.New("invalid argon2id parameters")
	}
	var err error
	if params.salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
		return params, errors.InternalServer.New("invalid argon2id salt")
	}
	if params.hash, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
		return params, errors.InternalServer.New("invalid argon2id hash")
	}
	if params.memory < minimalMemory || params.iterations == 0 || params.parallelism == 0 || len(params.hash) == 0 {
		return params, errors.InternalServer.New("argon2id parameters are too weak")
	}
	return params, nil
}

// checkPassword проверяет пароль с паузами после неудач. Адрес клиента за прокси не виден,
// поэтому счётчик общий: перебор может временно закрыть вход и владельцу, но не откроет его
func (s *OAuthService) checkPassword(password string, now time.Time) error {
	s.passwordLock.Lock()
	defer s.passwordLock.Unlock()

	if now.Before(s.lockedUntil) {
		return model.ErrTooManyAttempts
	}

	computed := argon2.IDKey([]byte(password), s.password.salt, s.password.iterations, s.password.memory, s.password.parallelism, uint32(len(s.password.hash))) //nolint:gosec
	if subtle.ConstantTimeCompare(computed, s.password.hash) == 1 {
		s.failedAttempts = 0
		return nil
	}

	// Каждая неудача сверх бесплатных удваивает паузу
	s.failedAttempts++
	if s.failedAttempts >= freeAttempts {
		lockout := firstLockout << min(s.failedAttempts-freeAttempts, 4)
		s.lockedUntil = now.Add(min(lockout, maximalLockout))
	}
	return model.ErrInvalidCredentials
}
