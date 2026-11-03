package repository

import (
	"os"
	"path/filepath"
	"server/internal/utils/errors"
)

const (
	directoryMode = 0o755 // Права новых папок
	noteMode      = 0o644 // Права заметок
)

// writeFileAtomically пишет через временный файл: при сбое заметка остаётся прежней
func writeFileAtomically(absolutePath string, content []byte) (string, error) {
	if len(content) > MaxNoteBytes {
		return "", errors.BadRequest.New("note exceeds size limit")
	}

	// Создаём папку и временный файл рядом с заметкой
	directory := filepath.Dir(absolutePath)
	if err := os.MkdirAll(directory, directoryMode); err != nil {
		return "", errors.InternalServer.Wrap(err)
	}
	temporary, err := os.CreateTemp(directory, ".tmp-*")
	if err != nil {
		return "", errors.InternalServer.Wrap(err)
	}
	defer os.Remove(temporary.Name())

	// Записываем содержимое на диск
	if _, err = temporary.Write(content); err != nil {
		temporary.Close()
		return "", errors.InternalServer.Wrap(err)
	}
	if err = temporary.Sync(); err != nil {
		temporary.Close()
		return "", errors.InternalServer.Wrap(err)
	}
	if err = temporary.Close(); err != nil {
		return "", errors.InternalServer.Wrap(err)
	}

	// Подменяем заметку готовым файлом
	if err = os.Chmod(temporary.Name(), noteMode); err != nil {
		return "", errors.InternalServer.Wrap(err)
	}
	if err = os.Rename(temporary.Name(), absolutePath); err != nil {
		return "", errors.InternalServer.Wrap(err)
	}
	return hashContent(content), nil
}
