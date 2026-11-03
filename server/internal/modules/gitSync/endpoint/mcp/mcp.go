// Package mcp — инструмент MCP для коммитов
package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/gitSync/model"
	gitSyncService "server/internal/modules/gitSync/service"
)

var _ GitSyncService = new(gitSyncService.GitSyncService)

// GitSyncService — интерфейс бизнес-логики коммитов
type GitSyncService interface {
	Commit(context.Context, model.CommitReq) (model.CommitResult, error)
}

// GitSyncTools — обработчики инструментов синхронизации
type GitSyncTools struct {
	gitSyncService GitSyncService // Сервис синхронизации
}

// Register добавляет инструмент коммита на MCP-сервер
func Register(server *sdk.Server, gitSyncService GitSyncService) {
	tools := &GitSyncTools{gitSyncService: gitSyncService}

	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "commit",
		Description: "Коммитит все накопленные правки. Отправляет их сервер сам",
	}, tools.Commit)
}
