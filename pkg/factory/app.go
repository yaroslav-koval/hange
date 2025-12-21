package factory

import (
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
	"github.com/yaroslav-koval/hange/pkg/crypt"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
	"github.com/yaroslav-koval/hange/pkg/git"
)

type AppBuilder interface {
	GetAuth() (auth.Auth, error)
	GetAIAgent() (agent.AIAgent, error)
	GetConfigurator() (config.Configurator, error)
	GetFileProvider() (fileprovider.FileProvider, error)
	GetGitChangesProvider() (git.ChangesProvider, error)
}

type AppFactory interface {
	CreateConfigurator() (config.Configurator, error)
	CreateTokenFetcher(config.Configurator) (auth.TokenFetcher, error)
	CreateTokenStorer(config.Configurator) (auth.TokenStorer, error)
	CreateBase64Encryptor() (crypt.Encryptor, error)
	CreateBase64Decryptor() (crypt.Decryptor, error)
	CreateFileProvider() (fileprovider.FileProvider, error)
	CreateGitChangesProvider() (git.ChangesProvider, error)
}

type AgentFactory interface {
	CreateCommitProcessor(auth.Auth) (agent.CommitProcessor, error)
	CreateExplainProcessor(auth.Auth) (agent.ExplainProcessor, error)
}

func NewAppBuilder(appFactory AppFactory, agentFactory AgentFactory) AppBuilder {
	return &lazyAppBuilder{
		appFactory:   appFactory,
		agentFactory: agentFactory,
		au:           newLazyInitializer[auth.Auth](),
		ag:           newLazyInitializer[agent.AIAgent](),
		cfg:          newLazyInitializer[config.Configurator](),
		fp:           newLazyInitializer[fileprovider.FileProvider](),
		gi:           newLazyInitializer[git.ChangesProvider](),
	}
}

type lazyAppBuilder struct {
	appFactory   AppFactory
	agentFactory AgentFactory

	au  *lazyInitializer[auth.Auth]
	ag  *lazyInitializer[agent.AIAgent]
	cfg *lazyInitializer[config.Configurator]
	fp  *lazyInitializer[fileprovider.FileProvider]
	gi  *lazyInitializer[git.ChangesProvider]
}

func (ab *lazyAppBuilder) GetAuth() (auth.Auth, error) {
	return ab.au.Get(func() (auth.Auth, error) {
		configurator, err := ab.GetConfigurator()
		if err != nil {
			return nil, err
		}

		tokenStorer, err := ab.appFactory.CreateTokenStorer(configurator)
		if err != nil {
			return nil, err
		}

		tokenFetcher, err := ab.appFactory.CreateTokenFetcher(configurator)
		if err != nil {
			return nil, err
		}

		encryptor, err := ab.appFactory.CreateBase64Encryptor()
		if err != nil {
			return nil, err
		}

		decryptor, err := ab.appFactory.CreateBase64Decryptor()
		if err != nil {
			return nil, err
		}

		return auth.NewAuth(
			tokenStorer,
			tokenFetcher,
			encryptor,
			decryptor,
		), nil
	})
}

func (ab *lazyAppBuilder) GetAIAgent() (agent.AIAgent, error) {
	return ab.ag.Get(func() (agent.AIAgent, error) {
		au, err := ab.GetAuth()
		if err != nil {
			return nil, err
		}

		cp, err := ab.agentFactory.CreateCommitProcessor(au)
		if err != nil {
			return nil, err
		}

		ep, err := ab.agentFactory.CreateExplainProcessor(au)
		if err != nil {
			return nil, err
		}

		return agent.NewAgent(cp, ep)
	})
}

func (ab *lazyAppBuilder) GetConfigurator() (config.Configurator, error) {
	return ab.cfg.Get(func() (config.Configurator, error) {
		return ab.appFactory.CreateConfigurator()
	})
}

func (ab *lazyAppBuilder) GetFileProvider() (fileprovider.FileProvider, error) {
	return ab.fp.Get(func() (fileprovider.FileProvider, error) {
		return ab.appFactory.CreateFileProvider()
	})
}

func (ab *lazyAppBuilder) GetGitChangesProvider() (git.ChangesProvider, error) {
	return ab.gi.Get(func() (git.ChangesProvider, error) {
		return ab.appFactory.CreateGitChangesProvider()
	})
}
