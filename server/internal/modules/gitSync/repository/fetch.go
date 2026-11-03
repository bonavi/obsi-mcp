package repository

import "context"

// Fetch загружает рабочую ветку
func (r *GitSyncRepository) Fetch(ctx context.Context) error {
	_, err := r.run(ctx, "fetch", "-q", r.config.Remote, r.config.Branch)
	return err
}
