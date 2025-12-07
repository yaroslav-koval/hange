package factory

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenfetch"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenstore"
	"github.com/yaroslav-koval/hange/pkg/config"
)

type App struct {
	Auth   auth.Auth
	Config config.Configurator
}

func NewCLIApp(configPath string) (App, error) {
	// TODO lazy loading. Initialize a component only when it's needed

	cliConfig, err := config.NewCLIConfig(configPath)
	if err != nil {
		return App{}, err
	}

	return App{
		Auth: auth.NewAuth(
			tokenstore.NewConfigTokenStorer(cliConfig),
			tokenfetch.NewConfigTokenFetcher(cliConfig),
		),
		Config: cliConfig,
	}, nil
}
