package repository

import (
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// isDenied проверяет путь от корня: скрытые компоненты и список запретов.
// Сравнение в NFC и без регистра, иначе NFD-имя или другой регистр обошли бы запрет
func (r *NoteRepository) isDenied(relativePath string) bool {
	relativePath = strings.ToLower(norm.NFC.String(filepath.ToSlash(relativePath)))
	if relativePath == "." || relativePath == "" {
		return false
	}

	// Скрытые файлы и папки недоступны всегда
	for _, part := range strings.Split(relativePath, "/") {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}

	// Проверяем список запретов
	for _, denyPath := range r.denyPaths {
		if relativePath == denyPath || strings.HasPrefix(relativePath, denyPath+"/") {
			return true
		}
	}
	return false
}
