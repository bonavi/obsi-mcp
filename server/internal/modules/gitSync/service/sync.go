package service

import (
	"context"
)

// sync подтягивает удалённую ветку и пушит локальные коммиты. При незакоммиченных
// правках ничего не делает: их заберёт следующий запуск после коммита
func (s *GitSyncService) sync(ctx context.Context) error {
	files, err := s.gitSyncRepository.GetDirtyFiles(ctx)
	if err != nil || len(files) > 0 {
		return err
	}
	if err = s.gitSyncRepository.Fetch(ctx); err != nil {
		return err
	}

	// Переносим локальные коммиты поверх удалённых
	if err = s.gitSyncRepository.Rebase(ctx); err != nil {
		return s.handleRebaseFailure(ctx, err)
	}

	// Пушим, если есть что
	hasCommits, err := s.gitSyncRepository.HasUnpushedCommits(ctx)
	if err != nil || !hasCommits {
		return err
	}
	return s.gitSyncRepository.PushCurrentBranch(ctx)
}
