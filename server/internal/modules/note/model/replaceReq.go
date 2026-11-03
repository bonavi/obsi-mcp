package model

// ReplaceReq — запрос замены текста
type ReplaceReq struct {
	Path         string // Путь
	OldText      string // Заменяемый текст
	NewText      string // Новый текст
	ExpectedHash string // Ожидаемый хэш
}
