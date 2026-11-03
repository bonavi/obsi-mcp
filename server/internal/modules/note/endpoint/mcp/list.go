package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// ListIn — аргументы list
type ListIn struct {
	Directory string `json:"dir,omitempty" jsonschema:"Папка от корня базы, нет — корень"`
}

// EntryView — элемент папки
type EntryView struct {
	Path        string `json:"path"`   // Путь от корня базы
	IsDirectory bool   `json:"is_dir"` // Признак папки
}

// ListOut — содержимое папки
type ListOut struct {
	Items       []EntryView `json:"items"`     // Элементы папки
	IsTruncated bool        `json:"truncated"` // Признак обрезанного списка
}

// List вызывает инструмент list
func (t *NoteTools) List(ctx context.Context, _ *sdk.CallToolRequest, in ListIn) (*sdk.CallToolResult, ListOut, error) {
	result, err := t.noteService.List(ctx, model.ListReq{Directory: in.Directory})
	if err != nil {
		return nil, ListOut{}, mcpTransport.ToolError(err)
	}
	return nil, newListOut(result), nil
}
