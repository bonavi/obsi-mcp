package repository

import (
	"bytes"
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
)

const noteChangedMessage = "note changed since it was read, read it again"

// Replace заменяет текст, который встречается в заметке ровно один раз
func (r *NoteRepository) Replace(req model.ReplaceReq) (string, error) {
	absolutePath, err := r.resolveWritablePath(req.Path)
	if err != nil {
		return "", err
	}
	current, err := readFile(absolutePath)
	if err != nil {
		return "", err
	}
	if hashContent(current) != req.ExpectedHash {
		return "", errors.Conflict.New(noteChangedMessage)
	}
	if req.OldText == "" {
		return "", errors.BadRequest.New("text to replace not found")
	}

	// Заменяем только однозначное вхождение
	switch bytes.Count(current, []byte(req.OldText)) {
	case 0:
		return "", errors.BadRequest.New("text to replace not found")
	case 1:
	default:
		return "", errors.BadRequest.New("text to replace is not unique, add more context")
	}
	return writeFileAtomically(absolutePath, bytes.Replace(current, []byte(req.OldText), []byte(req.NewText), 1))
}
