package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// MoveIn — аргументы move
type MoveIn struct {
	From string `json:"from" jsonschema:"Текущий путь заметки"`
	To   string `json:"to" jsonschema:"Новый путь, его не должно существовать"`
}

// Move вызывает инструмент move
func (t *NoteTools) Move(ctx context.Context, _ *sdk.CallToolRequest, in MoveIn) (*sdk.CallToolResult, ChangeOut, error) {
	result, err := t.noteService.Move(ctx, model.MoveReq{From: in.From, To: in.To})
	if err != nil {
		return nil, ChangeOut{}, mcpTransport.ToolError(err)
	}
	return nil, newChangeOut(result), nil
}
