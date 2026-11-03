package service

// WithinWriteLock выполняет функцию под блокировкой записи
func (s *GitSyncService) WithinWriteLock(call func() error) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return call()
}
