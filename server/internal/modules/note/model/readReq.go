package model

// ReadReq — запрос страницы заметки
type ReadReq struct {
	Path   string // Путь
	Offset int    // Смещение
}
