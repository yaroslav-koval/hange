package agentfactory

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/agent/commit"
	"github.com/yaroslav-koval/hange/pkg/agent/explain"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

func NewOpenAIFactory() factory.AgentFactory {
	return &openAIFactory{}
}

type openAIFactory struct {
}

func (o *openAIFactory) CreateCommitProcessor(auth auth.Auth) (agent.CommitProcessor, error) {
	c, err := o.createOpenAIClient(auth)
	if err != nil {
		return nil, err
	}

	return commit.NewOpenAICommitProcessor(c), nil
}

func (o *openAIFactory) CreateExplainProcessor(auth auth.Auth) (agent.ExplainProcessor, error) {
	c, err := o.createOpenAIClient(auth)
	if err != nil {
		return nil, err
	}

	return explain.NewOpenAIExplainProcessor(c), nil
}

func (o *openAIFactory) createOpenAIClient(auth auth.Auth) (*openai.Client, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}

	c := openai.NewClient(
		option.WithAPIKey(token),
	)

	return &c, nil
}
