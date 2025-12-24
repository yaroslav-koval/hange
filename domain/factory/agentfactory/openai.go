package agentfactory

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/yaroslav-koval/hange/domain/agent"
	"github.com/yaroslav-koval/hange/domain/agent/commit"
	"github.com/yaroslav-koval/hange/domain/agent/explain"
	"github.com/yaroslav-koval/hange/domain/auth"
	"github.com/yaroslav-koval/hange/domain/factory"
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
