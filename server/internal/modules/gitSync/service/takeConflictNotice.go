package service

// TakeConflictNotice возвращает сообщение о последнем конфликте один раз
func (s *GitSyncService) TakeConflictNotice() (notice string) {
	_ = s.WithinWriteLock(func() error {
		if s.conflictBranch != "" {
			notice = "Earlier edits conflicted with a newer remote version and were saved to branch " + s.conflictBranch + ", not the main branch. Tell the user."
			s.conflictBranch = ""
		}
		return nil
	})
	return notice
}
