package service

import (
	"context"
	"server/internal/modules/note/model"
)

// Search ищет текст по заметкам
func (s *NoteService) Search(_ context.Context, req model.SearchReq) (result model.SearchResult, err error) {
	err = s.gitSyncService.WithinReadLock(func() (err error) {
		result, err = s.noteRepository.Search(req)
		return err
	})
	return result, err
}
