package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// AppendIn — аргументы append
type AppendIn struct {
	Path string `json:"path" jsonschema:"Путь к заметке от корня базы"`
	Text string `json:"text" jsonschema:"Текст для добавления в конец"`
}

// Append вызывает инструмент append
func (t *NoteTools) Append(ctx context.Context, _ *sdk.CallToolRequest, in AppendIn) (*sdk.CallToolResult, ChangeOut, error) {
	result, err := t.noteService.Append(ctx, model.AppendReq{Path: in.Path, Text: in.Text})
	if err != nil {
		return nil, ChangeOut{}, mcpTransport.ToolError(err)
	}
	return nil, newChangeOut(result), nil
}
