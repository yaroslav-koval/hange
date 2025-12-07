package tokenstore

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
)

func NewConfigTokenStorer(config config.Configurator) auth.TokenStorer {
	return &configTokenStorer{
		config: config,
	}
}

type configTokenStorer struct {
	config config.Configurator
}

func (c *configTokenStorer) Store(token string) error {
	if err := c.config.WriteField(config.AuthTokenPath, token); err != nil {
		return err
	}

	return nil
}
