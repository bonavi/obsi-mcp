package model

// SearchResult — результат поиска
type SearchResult struct {
	Hits        []Hit // Найденные строки
	IsTruncated bool  // Признак обрезанного результата
}
