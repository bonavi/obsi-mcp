package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	mcpTransport "server/internal/utils/mcp"
)

// ReplaceIn — аргументы replace
type ReplaceIn struct {
	Path         string `json:"path" jsonschema:"Путь к заметке от корня базы"`
	OldText      string `json:"old_text" jsonschema:"Точный текст, который встречается один раз"`
	NewText      string `json:"new_text" jsonschema:"Новый текст"`
	ExpectedHash string `json:"expected_hash" jsonschema:"Хэш из read"`
}

// Replace вызывает инструмент replace
func (t *NoteTools) Replace(ctx context.Context, _ *sdk.CallToolRequest, in ReplaceIn) (*sdk.CallToolResult, ChangeOut, error) {
	result, err := t.noteService.Replace(ctx, model.ReplaceReq{Path: in.Path, OldText: in.OldText, NewText: in.NewText, ExpectedHash: in.ExpectedHash})
	if err != nil {
		return nil, ChangeOut{}, mcpTransport.ToolError(err)
	}
	return nil, newChangeOut(result), nil
}
