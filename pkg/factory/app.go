package factory

import (
	"github.com/yaroslav-koval/hange/pkg/agent"
	"github.com/yaroslav-koval/hange/pkg/agent/openaiagent"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
	"github.com/yaroslav-koval/hange/pkg/crypt"
)

type AppFactory interface {
	CreateConfigurator() (config.Configurator, error)
	CreateTokenFetcher(config.Configurator) (auth.TokenFetcher, error)
	CreateTokenStorer(config.Configurator) (auth.TokenStorer, error)
	CreateBase64Encryptor() (crypt.Encryptor, error)
	CreateBase64Decryptor() (crypt.Decryptor, error)
}

type App struct {
	Auth   auth.Auth
	Agent  agent.AIAgent
	Config config.Configurator
}

func BuildApp(appFactory AppFactory) (App, error) {
	configurator, err := appFactory.CreateConfigurator()
	if err != nil {
		return App{}, err
	}

	tokenStorer, err := appFactory.CreateTokenStorer(configurator)
	if err != nil {
		return App{}, err
	}

	tokenFetcher, err := appFactory.CreateTokenFetcher(configurator)
	if err != nil {
		return App{}, err
	}

	encryptor, err := appFactory.CreateBase64Encryptor()
	if err != nil {
		return App{}, err
	}

	decryptor, err := appFactory.CreateBase64Decryptor()
	if err != nil {
		return App{}, err
	}

	au := auth.NewAuth(
		tokenStorer,
		tokenFetcher,
		encryptor,
		decryptor,
	)

	ag, err := openaiagent.NewOpenAIAgent(au)
	if err != nil {
		return App{}, err
	}

	return App{
		Auth:   au,
		Agent:  ag,
		Config: configurator,
	}, nil
}
