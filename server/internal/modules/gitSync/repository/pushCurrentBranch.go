package repository

import "context"

// PushCurrentBranch отправляет HEAD в рабочую ветку
func (r *GitSyncRepository) PushCurrentBranch(ctx context.Context) error {
	return r.Push(ctx, r.config.Branch)
}
