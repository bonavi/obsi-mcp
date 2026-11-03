package repository

import (
	"context"
	"strings"
)

// HasUnpushedCommits проверяет, есть ли локальные коммиты, которых нет на удалённой ветке
func (r *GitSyncRepository) HasUnpushedCommits(ctx context.Context) (bool, error) {
	output, err := r.run(ctx, "rev-list", "--count", r.remoteBranch()+"..HEAD")
	return strings.TrimSpace(output) != "0", err
}
