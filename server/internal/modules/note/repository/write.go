package repository

import (
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
)

// Write создаёт заметку при пустом хэше, иначе заменяет именно ту версию
func (r *NoteRepository) Write(req model.WriteReq) (string, error) {
	absolutePath, err := r.resolveWritablePath(req.Path)
	if err != nil {
		return "", err
	}

	// Сверяем текущее состояние заметки с ожидаемым
	current, err := readFile(absolutePath)
	switch {
	case errors.CastError(err).ErrorType == errors.NotFound:
		if req.ExpectedHash != "" {
			return "", errors.Conflict.New(noteChangedMessage)
		}
	case err != nil:
		return "", err
	case req.ExpectedHash == "":
		return "", errors.Conflict.New("note already exists")
	case hashContent(current) != req.ExpectedHash:
		return "", errors.Conflict.New(noteChangedMessage)
	}
	return writeFileAtomically(absolutePath, []byte(req.Content))
}
