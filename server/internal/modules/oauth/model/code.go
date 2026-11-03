package model

import "time"

// Code — одноразовый код авторизации, живёт только в памяти
type Code struct {
	Value         string    // Значение кода
	ClientID      string    // Клиент, которому выдан код
	RedirectURI   string    // Адрес возврата
	CodeChallenge string    // PKCE challenge
	Resource      string    // Канонический URL MCP
	ExpiresAt     time.Time // Срок действия
}
