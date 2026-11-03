package service

import (
	"context"
	"pkg/log"
	"pkg/panicRecover"
	"time"
)

// Run периодически страхует незакоммиченное и синхронизирует клон
func (s *GitSyncService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()
	for {
		if err := s.worker(ctx); err != nil {
			log.LogError(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// worker выполняет один проход синхронизации
func (s *GitSyncService) worker(ctx context.Context) error {
	defer panicRecover.PanicRecover(func(err error) {
		log.Error(err)
	})

	return s.WithinWriteLock(func() error {
		if err := s.commitStale(ctx, time.Now()); err != nil {
			log.LogError(err)
		}
		return s.sync(ctx)
	})
}
