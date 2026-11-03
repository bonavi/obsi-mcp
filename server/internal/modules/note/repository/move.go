package repository

import (
	"io/fs"
	"os"
	"path/filepath"
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
)

// Move переносит заметку на свободный путь
func (r *NoteRepository) Move(req model.MoveReq) error {
	source, err := r.resolveWritablePath(req.From)
	if err != nil {
		return err
	}
	destination, err := r.resolveWritablePath(req.To)
	if err != nil {
		return err
	}
	if _, err = os.Stat(source); errors.Is(err, fs.ErrNotExist) {
		return errors.NotFound.New("note not found")
	}
	if _, err = os.Lstat(destination); err == nil {
		return errors.Conflict.New("note already exists")
	}
	if err = os.MkdirAll(filepath.Dir(destination), directoryMode); err != nil {
		return errors.InternalServer.Wrap(err)
	}
	if err = os.Rename(source, destination); err != nil {
		return errors.InternalServer.Wrap(err)
	}
	return nil
}
