package service

import (
	"context"
	"server/internal/modules/note/model"
)

// Read возвращает страницу заметки
func (s *NoteService) Read(_ context.Context, req model.ReadReq) (page model.Page, err error) {
	err = s.gitSyncService.WithinReadLock(func() (err error) {
		page, err = s.noteRepository.Read(req)
		return err
	})
	return page, err
}
