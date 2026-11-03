package repository

import (
	"path"
	"server/internal/utils/errors"
	"strings"
)

// resolveWritablePath разрешает путь к заметке, которую можно менять
func (r *NoteRepository) resolveWritablePath(relativePath string) (string, error) {
	if protectedNames[strings.ToLower(path.Base(relativePath))] {
		return "", errors.Forbidden.New(deniedPathMessage)
	}
	return r.resolveNotePath(relativePath)
}
