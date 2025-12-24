package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/yaroslav-koval/hange/domain/agent/entity"
	"github.com/yaroslav-koval/hange/domain/entities"
)

func NewAgent(cp CommitProcessor, ep ExplainProcessor) (AIAgent, error) {
	return &agent{
		cp: cp,
		ep: ep,
	}, nil
}

type agent struct {
	cp CommitProcessor
	ep ExplainProcessor
}

func (o *agent) ExplainFiles(ctx context.Context, files <-chan entities.File) (string, error) {
	defer o.ep.Cleanup(ctx)

	if err := o.ep.ProcessFiles(ctx, files); err != nil {
		return "", err
	}

	return o.ep.ExecuteExplainRequest(ctx)
}

func (o *agent) CreateCommitMessage(ctx context.Context, data entity.CommitData) (string, error) {
	if err := o.validateCommitParams(data); err != nil {
		return "", err
	}

	slog.Info("Commit data is sufficient. Waiting for LLM processing...")
	defer slog.Info("LLM finished processing")

	return o.cp.GenCommitMessage(ctx, data)
}

var ErrProvidedEmptyInput = errors.New("provided empty input")
var ErrNoStatusProvided = errors.New("either status or staged status should be provided")

func (o *agent) validateCommitParams(data entity.CommitData) error {
	if data.Status == "" && data.StagedStatus == "" {
		return ErrNoStatusProvided
	}

	if data.Status == "" {
		slog.Warn("Status is not provided, quality of response may be reduced")
	}

	if data.StagedStatus == "" {
		slog.Warn("Staged status is not provided, quality of response may be reduced")
	}

	if data.Diff == "" {
		return fmt.Errorf("%w: git diff. LLM can't construct commit message", ErrProvidedEmptyInput)
	}

	return nil
}
