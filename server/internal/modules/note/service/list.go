package service

import (
	"context"
	"server/internal/modules/note/model"
)

// List возвращает содержимое папки
func (s *NoteService) List(_ context.Context, req model.ListReq) (result model.ListResult, err error) {
	err = s.gitSyncService.WithinReadLock(func() (err error) {
		result, err = s.noteRepository.List(req)
		return err
	})
	return result, err
}
