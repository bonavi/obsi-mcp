package service

import (
	"context"
	"server/internal/modules/note/model"
)

// Write создаёт или заменяет заметку
func (s *NoteService) Write(_ context.Context, req model.WriteReq) (model.WriteResult, error) {
	return s.applyChange(func() (string, error) { return s.noteRepository.Write(req) })
}
