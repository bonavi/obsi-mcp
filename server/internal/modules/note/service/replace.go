package service

import (
	"context"
	"server/internal/modules/note/model"
)

// Replace заменяет текст в заметке
func (s *NoteService) Replace(_ context.Context, req model.ReplaceReq) (model.WriteResult, error) {
	return s.applyChange(func() (string, error) { return s.noteRepository.Replace(req) })
}
