package repository

import "server/internal/modules/note/model"

// Append дописывает текст в конец. Хэш не нужен: это не затирает чужую правку
func (r *NoteRepository) Append(req model.AppendReq) (string, error) {
	absolutePath, err := r.resolveWritablePath(req.Path)
	if err != nil {
		return "", err
	}
	current, err := readFile(absolutePath)
	if err != nil {
		return "", err
	}
	if len(current) > 0 && current[len(current)-1] != '\n' {
		current = append(current, '\n')
	}
	return writeFileAtomically(absolutePath, append(current, req.Text...))
}
