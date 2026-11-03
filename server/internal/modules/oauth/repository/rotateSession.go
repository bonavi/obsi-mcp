package repository

import (
	"server/internal/modules/oauth/model"
	"time"
)

// RotateSession одноразово меняет refresh-токен на новый. Повторное использование токена —
// признак кражи: отзывается вся цепочка. Возвращает false, если обмен не разрешён
func (r *OAuthRepository) RotateSession(tokenHash, newTokenHash string, expiresAt, now time.Time) (string, bool, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	session, found := r.sessions[tokenHash]
	if !found || now.After(session.ExpiresAt) {
		return "", false, nil
	}

	// Токен уже обменивали: отзываем цепочку
	if session.IsUsed {
		for hash, existing := range r.sessions {
			if existing.FamilyID == session.FamilyID {
				delete(r.sessions, hash)
			}
		}
		return "", false, r.persist(now)
	}

	// Обычная ротация
	session.IsUsed = true
	r.sessions[tokenHash] = session
	r.sessions[newTokenHash] = model.Session{FamilyID: session.FamilyID, IsUsed: false, ExpiresAt: expiresAt}
	return session.FamilyID, true, r.persist(now)
}
