package mcp

import (
	"pkg/slices"
	"server/internal/modules/note/model"
)

// newListOut конвертирует результат сервиса
func newListOut(result model.ListResult) ListOut {
	return ListOut{
		Items: slices.Map(result.Items, func(entry model.Entry) EntryView {
			return EntryView{Path: entry.Path, IsDirectory: entry.IsDirectory}
		}),
		IsTruncated: result.IsTruncated,
	}
}
