package repository

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"server/internal/utils/errors"
	"strings"
)

// run выполняет команду git и возвращает её вывод
func (r *GitSyncRepository) run(ctx context.Context, arguments ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	// Хуки отключены: файлы в клоне не должны влиять на то, что запускает сервер
	fullArguments := append([]string{"-c", "core.hooksPath=/dev/null", "-c", "core.quotePath=false"}, arguments...)
	command := exec.CommandContext(ctx, "git", fullArguments...) //nolint:gosec
	command.Dir = r.config.Directory
	command.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_NAME="+r.config.AuthorName, "GIT_AUTHOR_EMAIL="+r.config.AuthorEmail,
		"GIT_COMMITTER_NAME="+r.config.AuthorName, "GIT_COMMITTER_EMAIL="+r.config.AuthorEmail,
	)

	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		return output.String(), errors.InternalServer.Wrap(err).WithParams("command", arguments[0], "output", strings.TrimSpace(output.String()))
	}
	return output.String(), nil
}
