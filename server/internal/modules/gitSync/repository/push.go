package repository

import "context"

// Push отправляет HEAD в указанную ветку
func (r *GitSyncRepository) Push(ctx context.Context, branch string) error {
	_, err := r.run(ctx, "push", "-q", r.config.Remote, "HEAD:refs/heads/"+branch)
	return err
}
