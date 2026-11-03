package service

import (
	"context"
	"fmt"
	"time"
)

// commitStale коммитит изменения, если самая старая правка лежит дольше допустимого
func (s *GitSyncService) commitStale(ctx context.Context, now time.Time) error {
	files, err := s.gitSyncRepository.GetDirtyFiles(ctx)
	if err != nil {
		return err
	}

	// Ищем хотя бы один давно изменённый файл
	moment := now.Add(-s.autoCommitAfter)
	for _, file := range files {
		if s.gitSyncRepository.IsOlderThan(file, moment) {
			_, err = s.commitDirtyFiles(ctx, fmt.Sprintf("auto: %d файл(ов) без коммита", len(files)))
			return err
		}
	}
	return nil
}
