package repository

import "context"

// AbortRebase отменяет незавершённый rebase
func (r *GitSyncRepository) AbortRebase(ctx context.Context) error {
	_, err := r.run(ctx, "rebase", "--abort")
	return err
}
