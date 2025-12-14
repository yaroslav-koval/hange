package commit

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"github.com/yaroslav-koval/hange/pkg/agent"
)

func NewOpenAICommitProcessor(client *openai.Client) agent.CommitProcessor {
	return &commitProcessor{
		client: client,
	}
}

const commitModel = openai.ChatModelGPT5Nano

type commitProcessor struct {
	client *openai.Client
}

func (cp *commitProcessor) GenCommitMessage(ctx context.Context, data agent.CommitData) (string, error) {
	resp, err := cp.client.Responses.New(ctx, responses.ResponseNewParams{
		Instructions: openai.String(systemInstruction),
		Include: []responses.ResponseIncludable{
			responses.ResponseIncludableFileSearchCallResults,
		},
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(cp.buildInput(data)),
		},
		Model: commitModel,
	})
	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

const systemInstruction = `You write Git commit messages.

Hard requirements:

- Output EXACTLY ONE line of plain text.
- No quotes, no markdown, no code fences, no trailing period.
- Keep it short and specific (aim <= 72 chars).
- Summarize the net change across ALL files (what + why), using the diff and reason.`

func (cp *commitProcessor) buildInput(data agent.CommitData) string {
	b := strings.Builder{}

	if data.UserInput != "" {
		b.WriteString(fmt.Sprintf("User provided context:\n%s\n\n", data.UserInput))
	}

	if data.Status != "" {
		b.WriteString(fmt.Sprintf("GIT STATUS (porcelain):\n%s\n\n", data.Status))
	}

	if data.StagedStatus != "" {
		b.WriteString(fmt.Sprintf("GIT STAGED STATUS:\n%s\n\n", data.StagedStatus))
	}

	if data.Diff != "" {
		b.WriteString(fmt.Sprintf(`STAGED PATCH (unified diff):
<<<BEGIN PATCH>>>
%s
<<<END PATCH>>>\n\n`, data.Diff))
	}

	return b.String()
}
