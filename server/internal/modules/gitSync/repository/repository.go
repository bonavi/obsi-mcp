package repository

import (
	"server/internal/modules/gitSync/model"
	"time"
)

const commandTimeout = 2 * time.Minute // Предельное время команды git

// GitSyncRepository — команды git над клоном
type GitSyncRepository struct {
	config model.Config // Настройки клона
}

// NewGitSyncRepository создаёт репозиторий git
func NewGitSyncRepository(config model.Config) *GitSyncRepository {
	return &GitSyncRepository{config: config}
}
