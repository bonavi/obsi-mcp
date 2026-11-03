package model

// MoveReq — запрос переноса заметки
type MoveReq struct {
	From string // Текущий путь
	To   string // Новый путь
}
