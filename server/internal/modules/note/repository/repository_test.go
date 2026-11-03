package repository

import (
	"os"
	"path/filepath"
	pkgErrors "pkg/errors"
	"server/internal/modules/note/model"
	"server/internal/utils/errors"
	"strings"
	"testing"

	"golang.org/x/text/unicode/norm"
)

func setup(t *testing.T) (*NoteRepository, string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	put := func(rel, body string) {
		p := filepath.Join(root, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	put("a.md", "alpha секрета нет\n")
	put("Здоровье/x.md", "alpha тайна\n")
	put(".obsidian/app.md", "alpha\n")
	put("CLAUDE.md", "rules\n")
	put("pic.png", "bin")
	v, err := NewNoteRepository(root, []string{"Здоровье"})
	if err != nil {
		t.Fatal(err)
	}
	return v, root
}

// hasType проверяет тип кастомной ошибки
func hasType(err error, errorType pkgErrors.ErrorType) bool {
	return err != nil && pkgErrors.CastError(err).ErrorType == errorType
}

func TestDeniedPaths(t *testing.T) {
	v, root := setup(t)
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "s.md"), []byte("s"), 0o644)
	os.Symlink(filepath.Join(outside, "s.md"), filepath.Join(root, "link.md"))
	os.Symlink(outside, filepath.Join(root, "linkdir"))
	os.Symlink(filepath.Join(root, "Здоровье"), filepath.Join(root, "alias"))

	for _, p := range []string{
		"", "../x.md", "b/../../x.md", "b/../a.md", "/etc/passwd.md", "a\x00.md", `..\x.md`,
		".git/config.md", ".obsidian/app.md", "Здоровье/x.md", "здоровье/x.md",
		norm.NFD.String("Здоровье/x.md"), "./Здоровье/x.md", "Здоровье//x.md",
		"link.md", "linkdir/s.md", "alias/x.md", "pic.png",
	} {
		if _, err := v.Read(model.ReadReq{Path: p, Offset: 0}); !hasType(err, errors.Forbidden) {
			t.Errorf("Read(%q) = %v, want denied", p, err)
		}
	}
	for _, p := range []string{"Здоровье/new.md", "linkdir/new.md", "alias/new.md", "CLAUDE.md", "sub/agents.md", "x.sh", ".git/hooks/a.md"} {
		if _, err := v.Write(model.WriteReq{Path: p, Content: "x", ExpectedHash: ""}); !hasType(err, errors.Forbidden) {
			t.Errorf("Write(%q) = %v, want denied", p, err)
		}
	}
	if err := v.Move(model.MoveReq{From: "a.md", To: "Здоровье/a.md"}); !hasType(err, errors.Forbidden) {
		t.Errorf("Move into denied = %v", err)
	}
	if err := v.Move(model.MoveReq{From: "Здоровье/x.md", To: "b.md"}); !hasType(err, errors.Forbidden) {
		t.Errorf("Move from denied = %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "new.md")); err == nil {
		t.Error("write escaped vault")
	}
}

func TestDanglingSymlinkWrite(t *testing.T) {
	v, root := setup(t)
	target := filepath.Join(t.TempDir(), "evil.md")
	os.Symlink(target, filepath.Join(root, "dangling.md"))
	if _, err := v.Write(model.WriteReq{Path: "dangling.md", Content: "x", ExpectedHash: ""}); !hasType(err, errors.Forbidden) {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(target); err == nil {
		t.Fatal("write followed dangling symlink")
	}
}

func TestSearchAndListHideDenied(t *testing.T) {
	v, _ := setup(t)
	result, err := v.Search(model.SearchReq{Query: "ALPHA"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Hits) != 1 || result.Hits[0].Path != "a.md" {
		t.Fatalf("hits = %+v", result.Hits)
	}
	listing, err := v.List(model.ListReq{Directory: ""})
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range listing.Items {
		if it.Path == "Здоровье" || strings.HasPrefix(it.Path, ".") || it.Path == "pic.png" {
			t.Errorf("listed %q", it.Path)
		}
	}
	if _, err := v.List(model.ListReq{Directory: "Здоровье"}); !hasType(err, errors.Forbidden) {
		t.Errorf("List denied = %v", err)
	}
}

func TestWriteGuards(t *testing.T) {
	v, root := setup(t)
	if _, err := v.Write(model.WriteReq{Path: "a.md", Content: "x", ExpectedHash: ""}); !hasType(err, errors.Conflict) {
		t.Errorf("overwrite without hash = %v", err)
	}
	if _, err := v.Write(model.WriteReq{Path: "a.md", Content: "x", ExpectedHash: "stale"}); !hasType(err, errors.Conflict) {
		t.Errorf("stale hash = %v", err)
	}
	if _, err := v.Write(model.WriteReq{Path: "big.md", Content: strings.Repeat("x", MaxNoteBytes+1), ExpectedHash: ""}); !hasType(err, errors.BadRequest) {
		t.Errorf("big = %v", err)
	}
	h, err := v.Write(model.WriteReq{Path: "dir/new.md", Content: "one two two", ExpectedHash: ""})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.Replace(model.ReplaceReq{Path: "dir/new.md", OldText: "two", NewText: "x", ExpectedHash: h}); !hasType(err, errors.BadRequest) {
		t.Errorf("ambiguous = %v", err)
	}
	if _, err := v.Replace(model.ReplaceReq{Path: "dir/new.md", OldText: "zzz", NewText: "x", ExpectedHash: h}); !hasType(err, errors.BadRequest) {
		t.Errorf("no match = %v", err)
	}
	if _, err := v.Replace(model.ReplaceReq{Path: "dir/new.md", OldText: "one", NewText: "1", ExpectedHash: h}); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Append(model.AppendReq{Path: "dir/new.md", Text: "end\n"}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "dir/new.md"))
	if string(got) != "1 two two\nend\n" {
		t.Errorf("content %q", got)
	}
	left, _ := filepath.Glob(filepath.Join(root, "dir", ".tmp-*"))
	if len(left) != 0 {
		t.Errorf("temp files left: %v", left)
	}
}

func TestMove(t *testing.T) {
	v, root := setup(t)
	v.Write(model.WriteReq{Path: "b.md", Content: "b", ExpectedHash: ""})
	if err := v.Move(model.MoveReq{From: "a.md", To: "b.md"}); !hasType(err, errors.Conflict) {
		t.Errorf("over existing = %v", err)
	}
	if err := v.Move(model.MoveReq{From: "missing.md", To: "c.md"}); !hasType(err, errors.NotFound) {
		t.Errorf("missing = %v", err)
	}
	if err := v.Move(model.MoveReq{From: "a.md", To: "sub/c.md"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "sub/c.md")); err != nil {
		t.Error(err)
	}
}

func TestReadPagesKeepRunes(t *testing.T) {
	v, _ := setup(t)
	body := strings.Repeat("я", ReadPageBytes)
	if _, err := v.Write(model.WriteReq{Path: "long.md", Content: body, ExpectedHash: ""}); err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	offset, pages := 0, 0
	for {
		p, err := v.Read(model.ReadReq{Path: "long.md", Offset: offset})
		if err != nil {
			t.Fatal(err)
		}
		if p.Hash != hashContent([]byte(body)) {
			t.Fatal("page hash must cover the whole file")
		}
		sb.WriteString(p.Content)
		pages++
		if p.NextOffset == 0 {
			break
		}
		offset = p.NextOffset
	}
	if sb.String() != body || pages != 2 {
		t.Fatalf("pages=%d equal=%v", pages, sb.String() == body)
	}
}
