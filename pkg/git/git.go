package git

import "context"

type ChangesProvider interface {
	// Status of git, including staged and unstaged changes
	Status(context.Context) (string, error)
	// StagedStatus outputs general overview of files only for staged changes.
	StagedStatus(context.Context) (string, error)
	// StagedDiff outputs detailed information line by line file information about staged changes.
	// Second argument is a context scope. It's a number of lines outputted before and after lines with changes.
	// Bigger number means better context, smaller number means length optimized output.
	StagedDiff(ctx context.Context, linesAround int) (string, error)
}
