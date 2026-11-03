package service

import (
	"context"
	"server/internal/modules/note/model"
)

// Append дописывает текст в заметку
func (s *NoteService) Append(_ context.Context, req model.AppendReq) (model.WriteResult, error) {
	return s.applyChange(func() (string, error) { return s.noteRepository.Append(req) })
}
