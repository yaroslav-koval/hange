package agent

import (
	"context"

	"github.com/yaroslav-koval/hange/pkg/agent/entity"
	"github.com/yaroslav-koval/hange/pkg/entities"
)

type ExplainProcessor interface {
	ProcessFiles(context.Context, <-chan entities.File) error
	ExecuteExplainRequest(context.Context) (string, error)
	Cleanup(context.Context)
}

type CommitProcessor interface {
	GenCommitMessage(context.Context, entity.CommitData) (string, error)
}
