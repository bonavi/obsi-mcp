package service

import (
	"context"
	"server/internal/modules/gitSync/model"
)

// Commit коммитит все правки под блокировкой. Пушит их фоновая синхронизация
func (s *GitSyncService) Commit(ctx context.Context, req model.CommitReq) (result model.CommitResult, err error) {
	err = s.WithinWriteLock(func() (err error) {
		result.IsCommitted, err = s.commitDirtyFiles(ctx, req.Message)
		return err
	})
	result.Notice = s.TakeConflictNotice()
	return result, err
}
