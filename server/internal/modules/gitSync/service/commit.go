package service

import (
	"context"
	"pkg/log"
	"server/internal/modules/gitSync/model"
)

// Commit коммитит все правки и сразу отправляет их под блокировкой. Коммит уже сохранён
// в клоне, поэтому сбой отправки не ошибка: её повторит фоновая синхронизация
func (s *GitSyncService) Commit(ctx context.Context, req model.CommitReq) (result model.CommitResult, err error) {
	err = s.WithinWriteLock(func() (err error) {
		result.IsCommitted, err = s.commitDirtyFiles(ctx, req.Message)
		if err != nil {
			return err
		}

		// Обрыв запроса не должен убить git посреди rebase
		if err = s.sync(context.WithoutCancel(ctx)); err != nil {
			log.LogError(err)
			return nil
		}
		result.IsPushed = s.conflictBranch == ""
		return nil
	})
	result.Notice = s.TakeConflictNotice()
	return result, err
}
