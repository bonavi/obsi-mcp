package mcp

import "server/internal/modules/note/model"

// ChangeOut — результат изменения заметки
type ChangeOut struct {
	Hash   string `json:"hash,omitempty"`   // Новый хэш заметки
	Notice string `json:"notice,omitempty"` // Сообщение о конфликте git
}

// newChangeOut конвертирует результат сервиса
func newChangeOut(result model.WriteResult) ChangeOut {
	return ChangeOut{Hash: result.Hash, Notice: result.Notice}
}
