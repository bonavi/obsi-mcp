package repository

import "context"

// CommitAll коммитит все изменения
func (r *GitSyncRepository) CommitAll(ctx context.Context, message string) error {
	if _, err := r.run(ctx, "add", "-A"); err != nil {
		return err
	}
	_, err := r.run(ctx, "commit", "-q", "-m", message)
	return err
}
