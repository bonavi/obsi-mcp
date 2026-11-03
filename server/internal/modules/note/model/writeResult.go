package model

// WriteResult — результат изменения заметки
type WriteResult struct {
	Hash   string // Новый хэш заметки
	Notice string // Сообщение о конфликте git
}
