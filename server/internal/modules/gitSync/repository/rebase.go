package repository

import "context"

// Rebase переносит локальные коммиты поверх удалённой ветки
func (r *GitSyncRepository) Rebase(ctx context.Context) error {
	_, err := r.run(ctx, "rebase", "-q", r.remoteBranch())
	return err
}
