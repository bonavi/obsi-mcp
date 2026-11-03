package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"server/internal/utils/errors"
	"time"
)

const stateFileMode = 0o600 // Файл читает только сервер

// persist убирает просроченное и атомарно сохраняет сессии. Вызывается под блокировкой
func (r *OAuthRepository) persist(now time.Time) error {
	for hash, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			delete(r.sessions, hash)
		}
	}
	content, err := json.Marshal(r.sessions)
	if err != nil {
		return errors.InternalServer.Wrap(err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(r.stateFile), ".tmp-*")
	if err != nil {
		return errors.InternalServer.Wrap(err)
	}
	defer os.Remove(temporary.Name())
	if _, err = temporary.Write(content); err != nil {
		temporary.Close()
		return errors.InternalServer.Wrap(err)
	}
	if err = temporary.Close(); err != nil {
		return errors.InternalServer.Wrap(err)
	}
	if err = os.Chmod(temporary.Name(), stateFileMode); err != nil {
		return errors.InternalServer.Wrap(err)
	}
	if err = os.Rename(temporary.Name(), r.stateFile); err != nil {
		return errors.InternalServer.Wrap(err)
	}
	return nil
}
