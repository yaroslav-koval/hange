package tokenfetch

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/domain/config/consts"
	configurator_mock "github.com/yaroslav-koval/hange/mocks/configurator"
)

func TestFetch(t *testing.T) {
	t.Parallel()

	cfg := configurator_mock.NewMockConfigurator(t)

	fetcher := NewConfigTokenFetcher(cfg)

	cfg.EXPECT().ReadField(consts.AuthTokenPath).Return("token-value")

	token, err := fetcher.Fetch()
	require.NoError(t, err)
	require.Equal(t, "token-value", token)
}

func TestNilToken(t *testing.T) {
	t.Parallel()

	cfg := configurator_mock.NewMockConfigurator(t)

	fetcher := NewConfigTokenFetcher(cfg)

	cfg.EXPECT().ReadField(consts.AuthTokenPath).Return(nil)

	token, err := fetcher.Fetch()
	require.ErrorIs(t, err, ErrTokenNotSet)
	require.Empty(t, token)
}

func TestInvalidFormat(t *testing.T) {
	t.Parallel()

	cfg := configurator_mock.NewMockConfigurator(t)

	fetcher := NewConfigTokenFetcher(cfg)

	cfg.EXPECT().ReadField(consts.AuthTokenPath).Return(1234)

	token, err := fetcher.Fetch()
	require.ErrorIs(t, err, errInvalidFormat)
	require.Empty(t, token)
}
