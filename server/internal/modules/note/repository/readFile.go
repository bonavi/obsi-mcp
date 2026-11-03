package repository

import (
	"io/fs"
	"os"
	"server/internal/utils/errors"
)

// readFile читает файл заметки
func readFile(absolutePath string) ([]byte, error) {
	content, err := os.ReadFile(absolutePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errors.NotFound.New("note not found")
	}
	if err != nil {
		return nil, errors.InternalServer.Wrap(err)
	}
	return content, nil
}
