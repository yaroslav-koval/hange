package agent

type CommitData struct {
	// UserInput is a text helpful for LLM to understand context, like task description. Can be empty.
	UserInput string
	// git Status. All the files changed including not staged. Without file content
	Status string
	// StagedStatus is all the changed files that are staged. Without file content
	StagedStatus string
	// Diff is an actual representation of changes line by line.
	Diff string
}
