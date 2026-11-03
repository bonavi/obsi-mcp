package repository

import (
	"io/fs"
	"os"
	"path/filepath"
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Search ищет подстроку без учёта регистра в NFC, симлинки не обходит
func (r *NoteRepository) Search(req model.SearchReq) (model.SearchResult, error) {
	var result model.SearchResult
	needle := strings.ToLower(norm.NFC.String(req.Query))
	if needle == "" {
		return result, errors.BadRequest.New("empty query")
	}

	// Обходим базу, пропуская запрещённое
	err := filepath.WalkDir(r.root, func(currentPath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relativePath, err := filepath.Rel(r.root, currentPath)
		if err != nil {
			return err
		}
		if r.isDenied(relativePath) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !entry.Type().IsRegular() || !readableExtensions[strings.ToLower(filepath.Ext(currentPath))] {
			return nil
		}
		content, err := os.ReadFile(currentPath)
		if err != nil {
			return err
		}

		// Собираем подходящие строки
		for i, line := range strings.Split(string(content), "\n") {
			if !strings.Contains(strings.ToLower(norm.NFC.String(line)), needle) {
				continue
			}
			if len(result.Hits) == maxSearchHits {
				result.IsTruncated = true
				return filepath.SkipAll
			}
			result.Hits = append(result.Hits, model.Hit{Path: filepath.ToSlash(relativePath), Line: i + 1, Text: clipLine(line)})
		}
		return nil
	})
	if err != nil {
		return result, errors.InternalServer.Wrap(err)
	}
	return result, nil
}
