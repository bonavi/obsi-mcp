package model

import "time"

// Session — запись о refresh-токене. Сам токен не хранится, только его хэш
type Session struct {
	FamilyID  string    `json:"family_id"`  // Цепочка токенов одного входа
	IsUsed    bool      `json:"is_used"`    // Токен уже обменян
	ExpiresAt time.Time `json:"expires_at"` // Срок действия
}
