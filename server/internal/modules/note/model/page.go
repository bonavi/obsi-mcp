package model

// Page — страница заметки
type Page struct {
	Content    string // Текст страницы
	Hash       string // Хэш всего файла
	TotalBytes int    // Размер файла в байтах
	NextOffset int    // Смещение следующей страницы
}
