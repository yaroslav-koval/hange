package tokenstore

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	configurator_mock "github.com/yaroslav-koval/hange/mocks/configurator"
	"github.com/yaroslav-koval/hange/pkg/config/consts"
)

func TestStore(t *testing.T) {
	t.Parallel()

	cfg := configurator_mock.NewMockConfigurator(t)

	storer := NewConfigTokenStorer(cfg)

	cfg.EXPECT().WriteField(consts.AuthTokenPath, "token-value").Return(nil)

	err := storer.Store("token-value")
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	t.Parallel()

	cfg := configurator_mock.NewMockConfigurator(t)

	storer := NewConfigTokenStorer(cfg)

	expErr := errors.New("not valid token")
	cfg.EXPECT().WriteField(consts.AuthTokenPath, "token-value").Return(expErr)

	err := storer.Store("token-value")
	require.ErrorIs(t, err, expErr)
}
