package model

// WriteReq — запрос записи заметки
type WriteReq struct {
	Path         string // Путь
	Content      string // Текст
	ExpectedHash string // Ожидаемый хэш
}
