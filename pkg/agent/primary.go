package agent

import (
	"context"

	"github.com/yaroslav-koval/hange/pkg/agent/entity"
	"github.com/yaroslav-koval/hange/pkg/entities"
)

type AIAgent interface {
	// ExplainFiles takes a single file or a set of files and outputs explanation of them.
	// If folder is involved, files must have a relative (better option) or absolute path so LLM can see a folder structure.
	ExplainFiles(context.Context, <-chan entities.File) (string, error)
	// CreateCommitMessage receives context information and returns a commit message for git commit command.
	CreateCommitMessage(context.Context, entity.CommitData) (string, error)
}
