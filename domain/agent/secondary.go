package agent

import (
	"context"

	"github.com/yaroslav-koval/hange/domain/agent/entity"
	"github.com/yaroslav-koval/hange/domain/entities"
)

type ExplainProcessor interface {
	ProcessFiles(context.Context, <-chan entities.File) error
	ExecuteExplainRequest(context.Context) (string, error)
	Cleanup(context.Context)
}

type CommitProcessor interface {
	GenCommitMessage(context.Context, entity.CommitData) (string, error)
}
