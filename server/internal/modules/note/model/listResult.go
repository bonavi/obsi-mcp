package model

// ListResult — содержимое папки
type ListResult struct {
	Items       []Entry // Элементы папки
	IsTruncated bool    // Признак обрезанного списка
}
