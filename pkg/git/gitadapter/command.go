package gitadapter

import (
	"context"
	"fmt"
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
	commandExecutor commandExecutor
}

type commandExecutor interface {
	Execute(context.Context, string) (string, error)
}

const statusCommand = "git --no-pager status --porcelain"
const stagedStatusCommand = "git --no-pager diff --staged --no-color --stat"
const stagedDiffCommand = "git --no-pager diff --staged --no-color --no-ext-diff --patch --unified=%d"

func (g *gitChangesProvider) Status(ctx context.Context) (string, error) {
	return g.commandExecutor.Execute(ctx, statusCommand)
}

func (g *gitChangesProvider) StagedStatus(ctx context.Context) (string, error) {
	return g.commandExecutor.Execute(ctx, stagedStatusCommand)
}

func (g *gitChangesProvider) StagedDiff(ctx context.Context, linesAround int) (string, error) {
	strCmd := fmt.Sprintf(stagedDiffCommand, linesAround)

	return g.commandExecutor.Execute(ctx, strCmd)
}

type osExecutor struct{}

func (o *osExecutor) Execute(ctx context.Context, command string) (string, error) {
	cmd := o.stringToCommand(ctx, command)

	res, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func (o *osExecutor) stringToCommand(ctx context.Context, command string) *exec.Cmd {
	split := strings.Split(command, " ")

	if len(split) > 1 {
		return exec.CommandContext(ctx, split[0], split[1:]...)
	}

	return exec.CommandContext(ctx, split[0])
}
