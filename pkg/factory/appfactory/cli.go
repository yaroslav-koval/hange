package appfactory

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenfetch"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenstore"
	"github.com/yaroslav-koval/hange/pkg/config"
	"github.com/yaroslav-koval/hange/pkg/config/configcli"
	"github.com/yaroslav-koval/hange/pkg/crypt"
	"github.com/yaroslav-koval/hange/pkg/crypt/base64"
	"github.com/yaroslav-koval/hange/pkg/factory"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
	"github.com/yaroslav-koval/hange/pkg/fileprovider/directory"
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

func (c *cliFactory) CreateBase64Encryptor() (crypt.Encryptor, error) {
	return base64.NewBase64Encryptor(), nil
}

func (c *cliFactory) CreateBase64Decryptor() (crypt.Decryptor, error) {
	return base64.NewBase64Decryptor(), nil
}

func (c *cliFactory) CreateFileProvider() (fileprovider.FileProvider, error) {
	return directory.NewDirectoryFileProvider(), nil
}
