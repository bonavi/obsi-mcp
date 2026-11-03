package service

import (
	"context"
	"server/internal/modules/gitSync/repository"
	"sync"
	"time"
)

// Проверка реализации контракта репозитория
var _ GitSyncRepository = new(repository.GitSyncRepository)

// GitSyncRepository — интерфейс команд git
type GitSyncRepository interface {
	GetDirtyFiles(context.Context) ([]string, error)
	CommitAll(context.Context, string) error
	Fetch(context.Context) error
	Rebase(context.Context) error
	AbortRebase(context.Context) error
	HasUnmergedFiles(context.Context) (bool, error)
	HasUnpushedCommits(context.Context) (bool, error)
	Push(context.Context, string) error
	PushCurrentBranch(context.Context) error
	ResetToRemote(context.Context) error
	IsOlderThan(string, time.Time) bool
}

// GitSyncService — коммиты и синхронизация клона с удалённым репозиторием
type GitSyncService struct {
	gitSyncRepository GitSyncRepository // Команды git
	syncInterval      time.Duration     // Период синхронизации
	autoCommitAfter   time.Duration     // Возраст правки для автокоммита
	lock              sync.RWMutex      // Запись и git исключают чтение: rebase меняет файлы под читателем
	conflictBranch    string            // Ветка последнего конфликта, о котором ещё не сообщили
}

// NewGitSyncService создаёт сервис синхронизации
func NewGitSyncService(gitSyncRepository GitSyncRepository, syncInterval, autoCommitAfter time.Duration) *GitSyncService {
	return &GitSyncService{
		gitSyncRepository: gitSyncRepository,
		syncInterval:      syncInterval,
		autoCommitAfter:   autoCommitAfter,
		lock:              sync.RWMutex{},
		conflictBranch:    "",
	}
}
