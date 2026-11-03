package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// ReadIn — аргументы read
type ReadIn struct {
	Path   string `json:"path" jsonschema:"Путь к заметке от корня базы"`
	Offset int    `json:"offset,omitempty" jsonschema:"Смещение в байтах: next_offset предыдущей страницы"`
}

// ReadOut — страница заметки
type ReadOut struct {
	Content    string `json:"content"`               // Текст страницы
	Hash       string `json:"hash"`                  // Хэш всего файла
	TotalBytes int    `json:"total_bytes"`           // Размер файла в байтах
	NextOffset int    `json:"next_offset,omitempty"` // Смещение следующей страницы
}

// Read вызывает инструмент read
func (t *NoteTools) Read(ctx context.Context, _ *sdk.CallToolRequest, in ReadIn) (*sdk.CallToolResult, ReadOut, error) {
	result, err := t.noteService.Read(ctx, model.ReadReq{Path: in.Path, Offset: in.Offset})
	if err != nil {
		return nil, ReadOut{}, mcpTransport.ToolError(err)
	}
	return nil, ReadOut{Content: result.Content, Hash: result.Hash, TotalBytes: result.TotalBytes, NextOffset: result.NextOffset}, nil
}
