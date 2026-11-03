package repository

import (
	"context"
	"strings"
)

const minimalStatusEntryLength = 4 // Статус из двух букв, пробел и имя файла

// GetDirtyFiles возвращает файлы с незакоммиченными изменениями
func (r *GitSyncRepository) GetDirtyFiles(ctx context.Context) ([]string, error) {
	output, err := r.run(ctx, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}

	var files []string
	entries := strings.Split(output, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < minimalStatusEntryLength {
			continue
		}
		files = append(files, entry[3:])

		// У переименования следом идёт старое имя
		if entry[0] == 'R' || entry[0] == 'C' {
			i++
		}
	}
	return files, nil
}
