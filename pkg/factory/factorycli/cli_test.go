package factorycli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configurator_mock "github.com/yaroslav-koval/hange/mocks/configurator"
	"github.com/yaroslav-koval/hange/pkg/config/consts"
)

func TestNewCLIFactory_CreateConfigurator(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	cfgPath := filepath.Join(tempHome, ".hange")
	require.NoError(t, os.WriteFile(cfgPath, []byte("version: v1"), 0600))

	factory := NewCLIFactory(cfgPath)

	cfg, err := factory.CreateConfigurator()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestNewCLIFactory_CreateTokenStorer(t *testing.T) {
	t.Parallel()

	factory := NewCLIFactory("any")

	cfg := configurator_mock.NewMockConfigurator(t)
	cfg.EXPECT().WriteField(consts.AuthTokenPath, "token").Return(nil)

	storer, err := factory.CreateTokenStorer(cfg)
	require.NoError(t, err)
	require.NotNil(t, storer)

	assert.NoError(t, storer.Store("token"))
}

func TestNewCLIFactory_CreateTokenFetcher(t *testing.T) {
	t.Parallel()

	factory := NewCLIFactory("any")

	cfg := configurator_mock.NewMockConfigurator(t)
	cfg.EXPECT().ReadField(consts.AuthTokenPath).Return("token")

	fetcher, err := factory.CreateTokenFetcher(cfg)
	require.NoError(t, err)
	require.NotNil(t, fetcher)

	val, err := fetcher.Fetch()
	require.NoError(t, err)
	assert.Equal(t, "token", val)
}

func TestNewCLIFactory_CreateConfigurator_Error(t *testing.T) {
	t.Parallel()

	cfgPath := filepath.Join(t.TempDir(), "missing.yaml")
	factory := NewCLIFactory(cfgPath)

	cfg, err := factory.CreateConfigurator()
	require.Error(t, err)
	assert.Nil(t, cfg)
}
