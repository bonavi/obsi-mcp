package model

// Hit — найденная строка
type Hit struct {
	Path string // Путь от корня базы
	Line int    // Номер строки
	Text string // Текст строки
}
