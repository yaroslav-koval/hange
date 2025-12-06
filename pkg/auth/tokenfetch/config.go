package tokenfetch

import (
	"errors"

	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/config"
)

func NewConfigTokenFetcher() auth.TokenFetcher {
	return &configTokenFetcher{}
}

type configTokenFetcher struct{}

var errInvalidFormat = errors.New("invalid format of token, must be string")

func (c configTokenFetcher) Fetch() (string, error) {
	v, ok := config.ReadField(config.AuthTokenPath).(string)
	if !ok {
		return "", errInvalidFormat
	}

	return v, nil
}
