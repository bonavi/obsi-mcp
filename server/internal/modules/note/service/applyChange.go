package service

import "server/internal/modules/note/model"

// applyChange выполняет изменение под блокировкой и добавляет сообщение о конфликте
func (s *NoteService) applyChange(change func() (string, error)) (result model.WriteResult, err error) {
	err = s.gitSyncService.WithinWriteLock(func() (err error) {
		result.Hash, err = change()
		return err
	})
	result.Notice = s.gitSyncService.TakeConflictNotice()
	return result, err
}
