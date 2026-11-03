package model

// CommitResult — результат коммита
type CommitResult struct {
	IsCommitted bool   // Признак созданного коммита
	Notice      string // Сообщение о конфликте git
}
