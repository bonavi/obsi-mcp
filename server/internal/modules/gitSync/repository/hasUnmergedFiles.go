package repository

import "context"

// HasUnmergedFiles проверяет, есть ли файлы с конфликтом
func (r *GitSyncRepository) HasUnmergedFiles(ctx context.Context) (bool, error) {
	output, err := r.run(ctx, "ls-files", "--unmerged")
	return output != "", err
}
