package openaiagent

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/entities"
)

func NewOpenAIAgent(auth auth.Auth) (agent.AIAgent, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	client := openai.NewClient(
		option.WithAPIKey(token),
	)

	return &openAIUseCase{
		ep: newExplainProcessor(&client),
	}, nil
}

type ExplainProcessor interface {
	UploadFiles(ctx context.Context, files <-chan entities.File) error
	ExecuteExplainRequest(ctx context.Context) (string, error)
	Cleanup(ctx context.Context)
}

type openAIUseCase struct {
	auth auth.Auth
	ep   ExplainProcessor
}

func (o *openAIUseCase) ExplainFiles(ctx context.Context, files <-chan entities.File) (string, error) {
	defer o.ep.Cleanup(ctx)

	if err := o.ep.UploadFiles(ctx, files); err != nil {
		return "", err
	}

	return o.ep.ExecuteExplainRequest(ctx)
}
