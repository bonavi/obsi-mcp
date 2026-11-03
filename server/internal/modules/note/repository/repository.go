package repository

import (
	"path/filepath"
	"server/internal/utils/errors"
	"strings"

	"golang.org/x/text/unicode/norm"
)

const (
	MaxNoteBytes  = 1 << 20 // Максимальный размер заметки
	ReadPageBytes = 40000   // Размер страницы чтения
	maxListItems  = 1000    // Максимальное число элементов папки
	maxSearchHits = 50      // Максимальное число найденных строк
	maxLineBytes  = 300     // Максимальная длина строки в результате поиска
)

var (
	readableExtensions = map[string]bool{".md": true, ".canvas": true}         // Расширения, доступные через инструменты
	protectedNames     = map[string]bool{"claude.md": true, "agents.md": true} // Файлы с инструкциями для ИИ: их правка равна смене правил
)

// NoteRepository — доступ к файлам базы и все ограничения на пути
type NoteRepository struct {
	root      string   // Абсолютный путь к базе без симлинков
	denyPaths []string // Запрещённые папки в NFC и нижнем регистре
}

// NewNoteRepository создаёт репозиторий заметок
func NewNoteRepository(root string, denyPaths []string) (*NoteRepository, error) {

	// Раскрываем симлинки в корне, чтобы сравнивать с ним настоящие пути
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, errors.InternalServer.Wrap(err)
	}

	// Приводим запреты к форме для сравнения
	normalized := make([]string, 0, len(denyPaths))
	for _, denyPath := range denyPaths {
		denyPath = strings.Trim(norm.NFC.String(strings.TrimSpace(denyPath)), "/")
		if denyPath == "" {
			return nil, errors.InternalServer.New("empty deny path")
		}
		normalized = append(normalized, strings.ToLower(denyPath))
	}

	return &NoteRepository{root: realRoot, denyPaths: normalized}, nil
}
