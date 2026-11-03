package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// SearchIn — аргументы search
type SearchIn struct {
	Query string `json:"query" jsonschema:"Искомый текст без учёта регистра"`
}

// HitView — найденная строка
type HitView struct {
	Path string `json:"path"` // Путь от корня базы
	Line int    `json:"line"` // Номер строки
	Text string `json:"text"` // Текст строки
}

// SearchOut — результат поиска
type SearchOut struct {
	Hits        []HitView `json:"hits"`      // Найденные строки
	IsTruncated bool      `json:"truncated"` // Признак обрезанного результата
}

// Search вызывает инструмент search
func (t *NoteTools) Search(ctx context.Context, _ *sdk.CallToolRequest, in SearchIn) (*sdk.CallToolResult, SearchOut, error) {
	result, err := t.noteService.Search(ctx, model.SearchReq{Query: in.Query})
	if err != nil {
		return nil, SearchOut{}, mcpTransport.ToolError(err)
	}
	return nil, newSearchOut(result), nil
}
