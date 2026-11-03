// Package mcp — инструменты MCP для заметок. Удаления здесь намеренно нет
package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"server/internal/modules/note/model"
	noteService "server/internal/modules/note/service"
)

var _ NoteService = new(noteService.NoteService)

// NoteService — интерфейс бизнес-логики заметок
type NoteService interface {
	List(context.Context, model.ListReq) (model.ListResult, error)
	Read(context.Context, model.ReadReq) (model.Page, error)
	Search(context.Context, model.SearchReq) (model.SearchResult, error)
	Write(context.Context, model.WriteReq) (model.WriteResult, error)
	Replace(context.Context, model.ReplaceReq) (model.WriteResult, error)
	Append(context.Context, model.AppendReq) (model.WriteResult, error)
	Move(context.Context, model.MoveReq) (model.WriteResult, error)
}

// NoteTools — обработчики инструментов заметок
type NoteTools struct {
	noteService NoteService // Сервис заметок
}

// Register добавляет инструменты заметок на MCP-сервер
func Register(server *sdk.Server, noteService NoteService) {
	tools := &NoteTools{noteService: noteService}
	readOnly := &sdk.ToolAnnotations{ReadOnlyHint: true} //nolint:exhaustruct

	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "list",
		Description: "Возвращает заметки и папки в папке базы",
		Annotations: readOnly,
	}, tools.List)
	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "read",
		Description: "Читает заметку постранично. Возвращает hash, нужный для правки",
		Annotations: readOnly,
	}, tools.Read)
	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "search",
		Description: "Ищет текст по заметкам. Возвращает путь, номер строки и строку",
		Annotations: readOnly,
	}, tools.Search)
	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "write",
		Description: "Создает заметку или заменяет её целиком, если совпал expected_hash",
	}, tools.Write)
	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "replace",
		Description: "Заменяет текст, который встречается в заметке ровно один раз",
	}, tools.Replace)
	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "append",
		Description: "Дописывает текст в конец существующей заметки",
	}, tools.Append)
	sdk.AddTool(server, &sdk.Tool{ //nolint:exhaustruct
		Name:        "move",
		Description: "Переносит или переименовывает заметку. Ссылки на неё не обновляются",
	}, tools.Move)
}
