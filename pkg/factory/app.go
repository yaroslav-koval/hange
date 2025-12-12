package factory

import (
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/agent/openaiagent"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
	"github.com/yaroslav-koval/hange/pkg/crypt"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
)

type AppBuilder interface {
	BuildApp(appFactory AppFactory) (*App, error)
}

type AppFactory interface {
	CreateConfigurator() (config.Configurator, error)
	CreateTokenFetcher(config.Configurator) (auth.TokenFetcher, error)
	CreateTokenStorer(config.Configurator) (auth.TokenStorer, error)
	CreateBase64Encryptor() (crypt.Encryptor, error)
	CreateBase64Decryptor() (crypt.Decryptor, error)
	CreateFileProvider() (fileprovider.FileProvider, error)
}

type App struct {
	Auth         auth.Auth
	Agent        agent.AIAgent
	Config       config.Configurator
	FileProvider fileprovider.FileProvider
}

func NewAppBuilder() AppBuilder {
	return &appBuilder{}
}

type appBuilder struct{}

func (*appBuilder) BuildApp(appFactory AppFactory) (*App, error) {
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

	ag, err := openaiagent.NewOpenAIAgent(au)
	if err != nil {
		return nil, err
	}

	fp, err := appFactory.CreateFileProvider()
	if err != nil {
		return nil, err
	}

	return &App{
		Auth:         au,
		Agent:        ag,
		Config:       configurator,
		FileProvider: fp,
	}, nil
}
