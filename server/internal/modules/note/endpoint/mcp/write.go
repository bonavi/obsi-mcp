package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// WriteIn — аргументы write
type WriteIn struct {
	Path         string `json:"path" jsonschema:"Путь к заметке от корня базы"`
	Content      string `json:"content" jsonschema:"Полный текст заметки"`
	ExpectedHash string `json:"expected_hash,omitempty" jsonschema:"Хэш из read; нет — только для новой заметки"`
}

// Write вызывает инструмент write
func (t *NoteTools) Write(ctx context.Context, _ *sdk.CallToolRequest, in WriteIn) (*sdk.CallToolResult, ChangeOut, error) {
	result, err := t.noteService.Write(ctx, model.WriteReq{Path: in.Path, Content: in.Content, ExpectedHash: in.ExpectedHash})
	if err != nil {
		return nil, ChangeOut{}, mcpTransport.ToolError(err)
	}
	return nil, newChangeOut(result), nil
}
