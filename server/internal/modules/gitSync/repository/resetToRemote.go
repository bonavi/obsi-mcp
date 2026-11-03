package repository

import "context"

// ResetToRemote возвращает клон к состоянию удалённой ветки
func (r *GitSyncRepository) ResetToRemote(ctx context.Context) error {
	_, err := r.run(ctx, "reset", "-q", "--hard", r.remoteBranch())
	return err
}
