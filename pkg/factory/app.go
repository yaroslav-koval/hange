package factory

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
)

type AppFactory interface {
	CreateConfigurator() (config.Configurator, error)
	CreateTokenFetcher(config.Configurator) (auth.TokenFetcher, error)
	CreateTokenStorer(config.Configurator) (auth.TokenStorer, error)
}

type App struct {
	Auth   auth.Auth
	Config config.Configurator
}

func BuildApp(factory AppFactory) (App, error) {
	configurator, err := factory.CreateConfigurator()
	if err != nil {
		return App{}, err
	}

	tokenStorer, err := factory.CreateTokenStorer(configurator)
	if err != nil {
		return App{}, err
	}

	tokenFetcher, err := factory.CreateTokenFetcher(configurator)
	if err != nil {
		return App{}, err
	}

	return App{
		Auth:   auth.NewAuth(tokenStorer, tokenFetcher),
		Config: configurator,
	}, nil
}
