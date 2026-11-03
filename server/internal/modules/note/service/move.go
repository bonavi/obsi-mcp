package service

import (
	"context"
	"server/internal/modules/note/model"
)

// Move переносит заметку
func (s *NoteService) Move(_ context.Context, req model.MoveReq) (model.WriteResult, error) {
	return s.applyChange(func() (string, error) { return "", s.noteRepository.Move(req) })
}
