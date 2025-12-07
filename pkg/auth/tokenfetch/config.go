package tokenfetch

import (
	"errors"

	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
	"github.com/yaroslav-koval/hange/pkg/config/consts"
)

func NewConfigTokenFetcher(config config.Configurator) auth.TokenFetcher {
	return &configTokenFetcher{
		config: config,
	}
}

type configTokenFetcher struct {
	config config.Configurator
}

var errInvalidFormat = errors.New("invalid format of token, must be string")

func (c *configTokenFetcher) Fetch() (string, error) {
	v, ok := c.config.ReadField(consts.AuthTokenPath).(string)
	if !ok {
		return "", errInvalidFormat
	}

	return v, nil
}
