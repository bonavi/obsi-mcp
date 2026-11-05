package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/gitSync/model"
	mcpTransport "server/internal/utils/mcp"
)

// CommitIn — аргументы commit
type CommitIn struct {
	Message string `json:"message" jsonschema:"Краткое описание правок"`
}

// CommitOut — результат коммита
type CommitOut struct {
	IsCommitted bool   `json:"committed"`        // Признак созданного коммита
	IsPushed    bool   `json:"pushed"`           // Признак отправки в основную ветку
	Notice      string `json:"notice,omitempty"` // Сообщение о конфликте git
}

// Commit вызывает инструмент commit
func (t *GitSyncTools) Commit(ctx context.Context, _ *sdk.CallToolRequest, in CommitIn) (*sdk.CallToolResult, CommitOut, error) {
	result, err := t.gitSyncService.Commit(ctx, model.CommitReq{Message: in.Message})
	if err != nil {
		return nil, CommitOut{}, mcpTransport.ToolError(err)
	}
	return nil, CommitOut{IsCommitted: result.IsCommitted, IsPushed: result.IsPushed, Notice: result.Notice}, nil
}
