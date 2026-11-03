package repository

import (
	"server/internal/modules/oauth/model"
	"time"
)

// CreateCode сохраняет код и убирает просроченные
func (r *OAuthRepository) CreateCode(code model.Code, now time.Time) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for value, existing := range r.codes {
		if now.After(existing.ExpiresAt) {
			delete(r.codes, value)
		}
	}
	r.codes[code.Value] = code
}
