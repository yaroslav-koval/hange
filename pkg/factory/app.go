package factory

import (
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

	encryptor, err := factory.CreateBase64Encryptor()
	if err != nil {
		return App{}, err
	}

	decryptor, err := factory.CreateBase64Decryptor()
	if err != nil {
		return App{}, err
	}

	return App{
		Auth: auth.NewAuth(
			tokenStorer,
			tokenFetcher,
			encryptor,
			decryptor,
		),
		Config: configurator,
	}, nil
}
