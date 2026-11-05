package model

// CommitResult — результат коммита
type CommitResult struct {
	IsCommitted bool   // Признак созданного коммита
	IsPushed    bool   // Признак отправки в основную ветку
	Notice      string // Сообщение о конфликте git
}
