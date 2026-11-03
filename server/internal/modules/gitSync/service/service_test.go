package service

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"server/internal/modules/gitSync/model"
	"server/internal/modules/gitSync/repository"
)

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// setup: bare-репозиторий, клон сервера и клон «Мака».
func setup(t *testing.T) (*GitSyncService, string, string, string) {
	t.Helper()
	base := t.TempDir()
	bare, server, mac := filepath.Join(base, "bare.git"), filepath.Join(base, "server"), filepath.Join(base, "mac")
	run(t, base, "init", "-q", "--bare", "-b", "main", bare)
	run(t, base, "clone", "-q", bare, mac)
	write(t, mac, "a.md", "line\n")
	run(t, mac, "checkout", "-q", "-b", "main")
	run(t, mac, "add", "-A")
	run(t, mac, "commit", "-qm", "init")
	run(t, mac, "push", "-q", "origin", "main")
	run(t, base, "clone", "-q", "-b", "main", bare, server)
	config := model.Config{Directory: server, Remote: "origin", Branch: "main", AuthorName: "claude", AuthorEmail: "c@c"}
	return NewGitSyncService(repository.NewGitSyncRepository(config), time.Minute, 10*time.Minute), bare, server, mac
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCommitThenSyncPushes(t *testing.T) {
	r, bare, server, _ := setup(t)
	ctx := context.Background()
	if ok, err := r.commitDirtyFiles(ctx, "nothing"); err != nil || ok {
		t.Fatalf("empty commit: %v %v", ok, err)
	}
	write(t, server, "имя заметки.md", "x")
	if ok, err := r.commitDirtyFiles(ctx, "add\nnote"); err != nil || !ok {
		t.Fatalf("commit: %v %v", ok, err)
	}
	if err := r.sync(ctx); err != nil {
		t.Fatal(err)
	}
	if got := run(t, bare, "log", "-1", "--format=%s %an", "main"); got != "add note claude" {
		t.Fatalf("remote head %q", got)
	}
}

func TestSyncSkipsDirtyTree(t *testing.T) {
	r, bare, server, _ := setup(t)
	before := run(t, bare, "rev-parse", "main")
	write(t, server, "a.md", "dirty\n")
	if err := r.sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if run(t, bare, "rev-parse", "main") != before {
		t.Fatal("pushed dirty tree")
	}
}

func TestCommitStale(t *testing.T) {
	r, _, server, _ := setup(t)
	ctx := context.Background()
	write(t, server, "имя с пробелом.md", "x")
	if err := r.commitStale(ctx, time.Now()); err != nil {
		t.Fatalf("fresh: %v", err)
	}
	if got := run(t, server, "rev-list", "--count", "HEAD"); got != "1" {
		t.Fatalf("fresh edit committed, commits %s", got)
	}
	if err := r.commitStale(ctx, time.Now().Add(11*time.Minute)); err != nil {
		t.Fatalf("stale: %v", err)
	}
	if got := run(t, server, "log", "-1", "--format=%s"); !strings.HasPrefix(got, "auto:") {
		t.Fatalf("message %q", got)
	}
}

func TestRemoteChangesMerge(t *testing.T) {
	r, bare, server, mac := setup(t)
	ctx := context.Background()
	write(t, mac, "m.md", "mac")
	run(t, mac, "add", "-A")
	run(t, mac, "commit", "-qm", "mac")
	run(t, mac, "push", "-q", "origin", "main")
	write(t, server, "s.md", "server")
	r.commitDirtyFiles(ctx, "server")
	if err := r.sync(ctx); err != nil {
		t.Fatal(err)
	}
	if run(t, bare, "rev-list", "--count", "main") != "3" {
		t.Fatal("history not linear")
	}
	if r.TakeConflictNotice() != "" {
		t.Fatal("unexpected conflict")
	}
}

func TestConflictGoesToBranch(t *testing.T) {
	r, bare, server, mac := setup(t)
	ctx := context.Background()
	write(t, mac, "a.md", "mac\n")
	run(t, mac, "commit", "-qam", "mac")
	run(t, mac, "push", "-q", "origin", "main")
	macHead := run(t, mac, "rev-parse", "HEAD")

	write(t, server, "a.md", "server\n")
	r.commitDirtyFiles(ctx, "server edit")
	if err := r.sync(ctx); err != nil {
		t.Fatal(err)
	}
	notice := r.TakeConflictNotice()
	branch := run(t, bare, "for-each-ref", "--format=%(refname:short)", "refs/heads/claude")
	if !strings.Contains(notice, branch) {
		t.Fatalf("notice %q", notice)
	}
	if !strings.HasPrefix(branch, "claude/conflict-") {
		t.Fatalf("branch %q", branch)
	}
	if r.TakeConflictNotice() != "" {
		t.Fatal("conflict must be reported once")
	}
	if run(t, bare, "rev-parse", "main") != macHead {
		t.Fatal("main must stay at remote state")
	}
	if got := run(t, bare, "show", branch+":a.md"); got != "server" {
		t.Fatalf("edit lost: %q", got)
	}
	if run(t, server, "rev-parse", "HEAD") != macHead {
		t.Fatal("clone must return to remote")
	}
	if run(t, server, "status", "--porcelain") != "" {
		t.Fatal("clone must be clean")
	}
}

func TestConflictKeepsEditsWhenPushFails(t *testing.T) {
	r, bare, server, mac := setup(t)
	ctx := context.Background()
	write(t, mac, "a.md", "mac\n")
	run(t, mac, "commit", "-qam", "mac")
	run(t, mac, "push", "-q", "origin", "main")
	write(t, server, "a.md", "server\n")
	r.commitDirtyFiles(ctx, "server edit")
	head := run(t, server, "rev-parse", "HEAD")
	run(t, server, "fetch", "-q", "origin")
	run(t, server, "remote", "set-url", "--push", "origin", bare+"-missing")

	if err := r.sync(ctx); err == nil {
		t.Fatal("expected push error")
	}
	if run(t, server, "rev-parse", "HEAD") != head {
		t.Fatal("local commit lost")
	}
}

func TestRebaseFailureWithoutConflictKeepsEdits(t *testing.T) {
	r, bare, server, mac := setup(t)
	ctx := context.Background()
	write(t, mac, "m.md", "mac")
	run(t, mac, "add", "-A")
	run(t, mac, "commit", "-qm", "mac")
	run(t, mac, "push", "-q", "origin", "main")
	write(t, server, "s.md", "server")
	r.commitDirtyFiles(ctx, "server edit")
	head := run(t, server, "rev-parse", "HEAD")
	// Незаконченный rebase не даёт начать новый: сбой есть, конфликта нет.
	gitDir := filepath.Join(server, ".git", "rebase-merge")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := r.sync(ctx); err == nil {
		t.Fatal("expected rebase error")
	}
	if run(t, server, "rev-parse", "HEAD") != head {
		t.Fatal("local commit lost")
	}
	if r.TakeConflictNotice() != "" {
		t.Fatal("must not report a conflict")
	}
	if got := run(t, bare, "for-each-ref", "refs/heads/claude"); got != "" {
		t.Fatalf("unexpected branch %q", got)
	}
}
