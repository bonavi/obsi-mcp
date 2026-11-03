package repository

import (
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
	"unicode/utf8"
)

// Read отдаёт заметку страницами. Хэш всегда от всего файла, чтобы правка
// после постраничного чтения тоже была защищена от затирания
func (r *NoteRepository) Read(req model.ReadReq) (model.Page, error) {
	absolutePath, err := r.resolveNotePath(req.Path)
	if err != nil {
		return model.Page{}, err
	}
	content, err := readFile(absolutePath)
	if err != nil {
		return model.Page{}, err
	}
	if req.Offset < 0 || req.Offset > len(content) {
		return model.Page{}, errors.BadRequest.New("offset out of range")
	}

	// Определяем конец страницы, не разрывая символ
	end := req.Offset + ReadPageBytes
	if end >= len(content) {
		end = len(content)
	} else {
		for end > req.Offset && !utf8.RuneStart(content[end]) {
			end--
		}
	}

	page := model.Page{Content: string(content[req.Offset:end]), Hash: hashContent(content), TotalBytes: len(content), NextOffset: 0}
	if end < len(content) {
		page.NextOffset = end
	}
	return page, nil
}
