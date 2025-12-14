package gitadapter

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/yaroslav-koval/hange/pkg/git"
)

func NewGitChangesProvider() git.ChangesProvider {
	return &gitChangesProvider{
		commandExecutor: &osExecutor{},
	}
}

type gitChangesProvider struct {
	commandExecutor CommandExecutor
}

type CommandExecutor interface {
	Output(context.Context, string, ...string) (string, error)
	Run(context.Context, string, ...string) error
}

const statusCommand = "git --no-pager status --porcelain"
const stagedStatusCommand = "git --no-pager diff --staged --no-color --stat"
const stagedDiffCommand = "git --no-pager diff --staged --no-color --no-ext-diff --patch --unified=%d"

func (g *gitChangesProvider) Status(ctx context.Context) (string, error) {
	split := strings.Split(statusCommand, " ")

	return g.commandExecutor.Output(ctx, split[0], split[1:]...)
}

func (g *gitChangesProvider) StagedStatus(ctx context.Context) (string, error) {
	split := strings.Split(stagedStatusCommand, " ")

	return g.commandExecutor.Output(ctx, split[0], split[1:]...)
}

func (g *gitChangesProvider) StagedDiff(ctx context.Context, linesAround int) (string, error) {
	strCmd := fmt.Sprintf(stagedDiffCommand, linesAround)

	split := strings.Split(strCmd, " ")

	return g.commandExecutor.Output(ctx, split[0], split[1:]...)
}

func (g *gitChangesProvider) Commit(ctx context.Context, message string) error {
	return g.commandExecutor.Run(ctx, "git", []string{
		"--no-pager",
		"commit",
		"-m",
		message,
	}...)
}

type osExecutor struct{}

func (o *osExecutor) Output(ctx context.Context, command string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	slog.Debug(fmt.Sprintf("Executing command: %s", cmd.String()))

	res, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func (o *osExecutor) Run(ctx context.Context, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)

	slog.Debug(fmt.Sprintf("Executing command: %s", cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:\n%s\n", err, out)
	}

	return nil
}
