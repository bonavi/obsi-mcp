package service

// WithinReadLock выполняет функцию под блокировкой чтения
func (s *GitSyncService) WithinReadLock(call func() error) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return call()
}
