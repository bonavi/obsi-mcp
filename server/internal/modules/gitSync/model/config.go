package model

// Config — настройки работы с клоном
type Config struct {
	Directory   string // Папка с клоном
	Remote      string // Имя удалённого репозитория
	Branch      string // Рабочая ветка
	AuthorName  string // Имя автора коммитов
	AuthorEmail string // Почта автора коммитов
}
