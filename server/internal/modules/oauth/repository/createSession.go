package repository

import (
	"server/internal/modules/oauth/model"
	"time"
)

// CreateSession сохраняет новую сессию
func (r *OAuthRepository) CreateSession(tokenHash string, session model.Session, now time.Time) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.sessions[tokenHash] = session
	return r.persist(now)
}
