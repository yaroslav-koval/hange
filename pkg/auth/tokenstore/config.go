package tokenstore

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
)

func NewConfigTokenStorer() auth.TokenStorer {
	return &configTokenStorer{}
}

type configTokenStorer struct{}

func (c *configTokenStorer) Store(token string) error {
	if err := config.WriteField(config.AuthTokenPath, token); err != nil {
		return err
	}

	return nil
}
