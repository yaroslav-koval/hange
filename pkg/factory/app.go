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
	BuildApp(AppFactory, AgentFactory) (*App, error)
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

type App struct {
	Auth         auth.Auth
	Agent        agent.AIAgent
	Config       config.Configurator
	FileProvider fileprovider.FileProvider
	Git          git.ChangesProvider
}

func NewAppBuilder() AppBuilder {
	return &appBuilder{}
}

type appBuilder struct{}

func (*appBuilder) BuildApp(appFactory AppFactory, ab AgentFactory) (*App, error) {
	configurator, err := appFactory.CreateConfigurator()
	if err != nil {
		return nil, err
	}

	tokenStorer, err := appFactory.CreateTokenStorer(configurator)
	if err != nil {
		return nil, err
	}

	tokenFetcher, err := appFactory.CreateTokenFetcher(configurator)
	if err != nil {
		return nil, err
	}

	encryptor, err := appFactory.CreateBase64Encryptor()
	if err != nil {
		return nil, err
	}

	decryptor, err := appFactory.CreateBase64Decryptor()
	if err != nil {
		return nil, err
	}

	au := auth.NewAuth(
		tokenStorer,
		tokenFetcher,
		encryptor,
		decryptor,
	)

	fp, err := appFactory.CreateFileProvider()
	if err != nil {
		return nil, err
	}

	cp, err := ab.CreateCommitProcessor(au)
	if err != nil {
		return nil, err
	}

	ep, err := ab.CreateExplainProcessor(au)
	if err != nil {
		return nil, err
	}

	ag, err := agent.NewAgent(cp, ep)
	if err != nil {
		return nil, err
	}

	gi, err := appFactory.CreateGitChangesProvider()
	if err != nil {
		return nil, err
	}

	return &App{
		Auth:         au,
		Agent:        ag,
		Config:       configurator,
		FileProvider: fp,
		Git:          gi,
	}, nil
}
