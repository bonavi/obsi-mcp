package service

import (
	"context"
	"server/internal/utils/errors"
	"strings"
)

// commitDirtyFiles коммитит изменения. Возвращает false, если коммитить нечего
func (s *GitSyncService) commitDirtyFiles(ctx context.Context, message string) (bool, error) {

	// Сообщение коммита всегда в одну строку
	message = strings.Join(strings.Fields(message), " ")
	if message == "" {
		return false, errors.BadRequest.New("empty commit message")
	}

	files, err := s.gitSyncRepository.GetDirtyFiles(ctx)
	if err != nil || len(files) == 0 {
		return false, err
	}
	return true, s.gitSyncRepository.CommitAll(ctx, message)
}
