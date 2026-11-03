package repository

import (
	"encoding/json"
	"io/fs"
	"os"
	"server/internal/modules/oauth/model"
	"server/internal/utils/errors"
	"sync"
)

// OAuthRepository — коды в памяти и refresh-сессии в файле на отдельном томе
type OAuthRepository struct {
	lock      sync.Mutex               // Защита кодов и сессий
	stateFile string                   // Файл с сессиями
	codes     map[string]model.Code    // Коды по значению
	sessions  map[string]model.Session // Сессии по хэшу токена
}

// NewOAuthRepository загружает сессии. Повреждённый файл — ошибка старта, а не тихий сброс
func NewOAuthRepository(stateFile string) (*OAuthRepository, error) {
	sessions := map[string]model.Session{}
	content, err := os.ReadFile(stateFile)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return nil, errors.InternalServer.Wrap(err)
	default:
		if err = json.Unmarshal(content, &sessions); err != nil {
			return nil, errors.InternalServer.Wrap(err)
		}
	}
	return &OAuthRepository{lock: sync.Mutex{}, stateFile: stateFile, codes: map[string]model.Code{}, sessions: sessions}, nil
}
