package repository

import (
	"os"
	"path/filepath"
	"time"
)

// IsOlderThan проверяет, что файл изменён раньше указанного времени.
// У удалённого файла времени правки нет, поэтому он считается старым
func (r *GitSyncRepository) IsOlderThan(file string, moment time.Time) bool {
	info, err := os.Lstat(filepath.Join(r.config.Directory, file))
	return err != nil || !info.ModTime().After(moment)
}
