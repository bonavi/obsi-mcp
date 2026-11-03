package repository

import (
	"server/internal/modules/oauth/model"
	"time"
)

// ConsumeCode забирает код один раз. Просроченный код считается отсутствующим
func (r *OAuthRepository) ConsumeCode(value string, now time.Time) (model.Code, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	code, found := r.codes[value]
	delete(r.codes, value)
	if !found || now.After(code.ExpiresAt) {
		return model.Code{}, false
	}
	return code, true
}
