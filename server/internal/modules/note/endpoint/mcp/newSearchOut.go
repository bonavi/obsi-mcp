package mcp

import (
	"pkg/slices"
	"server/internal/modules/note/model"
)

// newSearchOut конвертирует результат сервиса
func newSearchOut(result model.SearchResult) SearchOut {
	return SearchOut{
		Hits: slices.Map(result.Hits, func(hit model.Hit) HitView {
			return HitView{Path: hit.Path, Line: hit.Line, Text: hit.Text}
		}),
		IsTruncated: result.IsTruncated,
	}
}
