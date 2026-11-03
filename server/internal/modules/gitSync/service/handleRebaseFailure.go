package service

import (
	"context"
	"time"
)

const conflictBranchPrefix = "claude/conflict-" // Префикс веток с конфликтующими правками

// handleRebaseFailure отводит правки в отдельную ветку только при настоящем конфликте.
// Любой другой сбой (сеть, диск, права) не должен уводить правки из рабочей ветки
func (s *GitSyncService) handleRebaseFailure(ctx context.Context, rebaseError error) error {
	hasConflict, unmergedError := s.gitSyncRepository.HasUnmergedFiles(ctx)
	abortError := s.gitSyncRepository.AbortRebase(ctx)
	if unmergedError != nil || !hasConflict {
		return rebaseError
	}
	if abortError != nil {
		return abortError
	}

	// Откат клона только после успешного пуша, иначе правки потеряются
	branch := conflictBranchPrefix + time.Now().UTC().Format("20060102-150405")
	if err := s.gitSyncRepository.Push(ctx, branch); err != nil {
		return err
	}
	if err := s.gitSyncRepository.ResetToRemote(ctx); err != nil {
		return err
	}
	s.conflictBranch = branch
	return nil
}
