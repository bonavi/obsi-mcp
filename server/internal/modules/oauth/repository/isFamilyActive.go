package repository

import "time"

// IsFamilyActive проверяет, что у цепочки есть действующий токен. Отзыв цепочки сразу закрывает и access-токены
func (r *OAuthRepository) IsFamilyActive(familyID string, now time.Time) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, session := range r.sessions {
		if session.FamilyID == familyID && !session.IsUsed && now.Before(session.ExpiresAt) {
			return true
		}
	}
	return false
}
