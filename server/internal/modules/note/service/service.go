package service

import (
	gitSyncService "server/internal/modules/gitSync/service"
	"server/internal/modules/note/model"
	"server/internal/modules/note/repository"
)

// Проверка реализации контракта репозитория
var _ NoteRepository = new(repository.NoteRepository)

// NoteRepository — интерфейс репозитория заметок
type NoteRepository interface {
	List(model.ListReq) (model.ListResult, error)
	Read(model.ReadReq) (model.Page, error)
	Search(model.SearchReq) (model.SearchResult, error)
	Write(model.WriteReq) (string, error)
	Replace(model.ReplaceReq) (string, error)
	Append(model.AppendReq) (string, error)
	Move(model.MoveReq) error
}

// Проверка реализации контракта сервиса синхронизации
var _ GitSyncService = new(gitSyncService.GitSyncService)

// GitSyncService — блокировка клона и сообщения о конфликтах
type GitSyncService interface {
	WithinReadLock(func() error) error
	WithinWriteLock(func() error) error
	TakeConflictNotice() string
}

// NoteService — сервис работы с заметками
type NoteService struct {
	noteRepository NoteRepository // Репозиторий заметок
	gitSyncService GitSyncService // Сервис синхронизации
}

// NewNoteService создаёт сервис заметок
func NewNoteService(noteRepository NoteRepository, gitSyncService GitSyncService) *NoteService {
	return &NoteService{noteRepository: noteRepository, gitSyncService: gitSyncService}
}
