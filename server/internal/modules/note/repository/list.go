package repository

import (
	"io/fs"
	"os"
	"path"
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
	"sort"
	"strings"
)

// List возвращает содержимое папки без запрещённых элементов
func (r *NoteRepository) List(req model.ListReq) (model.ListResult, error) {
	var result model.ListResult

	// Определяем папку
	absolutePath, base := r.root, ""
	if req.Directory != "" && req.Directory != "." {
		var err error
		if absolutePath, err = r.resolvePath(req.Directory); err != nil {
			return result, err
		}
		base = path.Clean(req.Directory)
	}

	// Читаем папку
	items, err := os.ReadDir(absolutePath)
	if errors.Is(err, fs.ErrNotExist) {
		return result, errors.NotFound.New("folder not found")
	}
	if err != nil {
		return result, errors.InternalServer.Wrap(err)
	}

	// Оставляем только доступное. Симлинки не показываем: они могут вести за пределы базы
	for _, item := range items {
		relativePath := path.Join(base, item.Name())
		if r.isDenied(relativePath) || item.Type()&fs.ModeSymlink != 0 {
			continue
		}
		if !item.IsDir() && !readableExtensions[strings.ToLower(path.Ext(relativePath))] {
			continue
		}
		result.Items = append(result.Items, model.Entry{Path: relativePath, IsDirectory: item.IsDir()})
	}
	sort.Slice(result.Items, func(i, j int) bool { return result.Items[i].Path < result.Items[j].Path })

	// Ограничиваем размер ответа
	if len(result.Items) > maxListItems {
		result.Items, result.IsTruncated = result.Items[:maxListItems], true
	}
	return result, nil
}
