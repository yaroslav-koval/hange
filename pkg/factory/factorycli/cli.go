package factorycli

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenfetch"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenstore"
	"github.com/yaroslav-koval/hange/pkg/config"
	"github.com/yaroslav-koval/hange/pkg/config/configcli"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

func NewCLIFactory(configPath string) factory.AppFactory {
	return &cliFactory{
		configPath: configPath,
	}
}

type cliFactory struct {
	configPath string
}

func (c *cliFactory) CreateConfigurator() (config.Configurator, error) {
	return configcli.NewCLIConfig(c.configPath)
}

func (c *cliFactory) CreateTokenFetcher(configurator config.Configurator) (auth.TokenFetcher, error) {
	return tokenfetch.NewConfigTokenFetcher(configurator), nil
}

func (c *cliFactory) CreateTokenStorer(configurator config.Configurator) (auth.TokenStorer, error) {
	return tokenstore.NewConfigTokenStorer(configurator), nil
}
