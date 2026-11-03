package repository

import (
	"path"
	"server/internal/utils/errors"
	"strings"
)

// resolveNotePath разрешает путь к файлу заметки
func (r *NoteRepository) resolveNotePath(relativePath string) (string, error) {
	if !readableExtensions[strings.ToLower(path.Ext(relativePath))] {
		return "", errors.Forbidden.New(deniedPathMessage)
	}
	return r.resolvePath(relativePath)
}
