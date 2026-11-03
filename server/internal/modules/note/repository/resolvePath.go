package repository

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"server/internal/utils/errors"
	"strings"
)

const deniedPathMessage = "path is not allowed"

// resolvePath превращает путь от клиента в абсолютный. Симлинки раскрываются до
// ближайшего существующего родителя, и запреты проверяются и на результате
func (r *NoteRepository) resolvePath(relativePath string) (string, error) {

	// Отсекаем заведомо недопустимые пути
	if relativePath == "" || strings.ContainsRune(relativePath, 0) || strings.Contains(relativePath, `\`) || path.IsAbs(relativePath) {
		return "", errors.Forbidden.New(deniedPathMessage)
	}
	for _, part := range strings.Split(relativePath, "/") {
		if part == ".." {
			return "", errors.Forbidden.New(deniedPathMessage)
		}
	}
	clean := path.Clean(relativePath)
	if r.isDenied(clean) {
		return "", errors.Forbidden.New(deniedPathMessage)
	}
	absolute := filepath.Join(r.root, filepath.FromSlash(clean))

	// Ищем ближайшего существующего родителя и раскрываем его симлинки
	existing, rest := absolute, ""
	for {
		real, err := filepath.EvalSymlinks(existing)
		if err == nil {
			existing = real
			break
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", errors.Forbidden.New(deniedPathMessage)
		}

		// Висячий симлинк нельзя считать отсутствующим файлом: запись пошла бы по нему
		if _, lstatErr := os.Lstat(existing); lstatErr == nil {
			return "", errors.Forbidden.New(deniedPathMessage)
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return "", errors.Forbidden.New(deniedPathMessage)
		}
		rest = filepath.Join(filepath.Base(existing), rest)
		existing = parent
	}

	// Проверяем, что настоящий путь остался в базе и не запрещён
	realRelative, err := filepath.Rel(r.root, existing)
	if err != nil || realRelative == ".." || strings.HasPrefix(realRelative, ".."+string(filepath.Separator)) {
		return "", errors.Forbidden.New(deniedPathMessage)
	}
	if r.isDenied(realRelative) {
		return "", errors.Forbidden.New(deniedPathMessage)
	}
	return filepath.Join(existing, rest), nil
}
