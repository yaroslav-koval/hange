package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	appfactory_mock "github.com/yaroslav-koval/hange/mocks/appfactory"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

func TestRootCommandConfiguration(t *testing.T) {
	require.Equal(t, "hange", rootCmd.Use)

	flag := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, flag)
	require.Equal(t, "config", flag.Name)
}

func TestPersistentPreRunESetsAppContext(t *testing.T) {
	originalPreRun := rootCmd.PersistentPreRunE
	t.Cleanup(func() { rootCmd.PersistentPreRunE = originalPreRun })

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		mockFactory := appfactory_mock.NewMockAppFactory(t)

		mockFactory.EXPECT().CreateConfigurator().Return(stubConfig{}, nil)
		mockFactory.EXPECT().CreateTokenFetcher(mock.Anything).Return(stubTokenFetcher{}, nil)
		mockFactory.EXPECT().CreateTokenStorer(mock.Anything).Return(stubTokenStorer{}, nil)
		mockFactory.EXPECT().CreateBase64Encryptor().Return(stubEncryptor{}, nil)
		mockFactory.EXPECT().CreateBase64Decryptor().Return(stubDecryptor{}, nil)

		app, err := factory.BuildApp(mockFactory)
		require.NoError(t, err)

		cmd.SetContext(appToCmdContext(cmd, &app))
		return nil
	}

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)

	val := cmd.Context().Value(appContextKey)
	require.IsType(t, &factory.App{}, val)
	app := val.(*factory.App)
	require.NotNil(t, app.Auth)
	require.NotNil(t, app.Config)
}

func TestPersistentPreRunEReturnsErrorWhenConfigMissing(t *testing.T) {
	missingCfg := filepath.Join(t.TempDir(), "missing.yaml")

	originalCfgFile := cfgPath
	t.Cleanup(func() { cfgPath = originalCfgFile })
	cfgPath = missingCfg

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
}

func writeTempConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	cfg := filepath.Join(dir, ".hange.yaml")
	err := os.WriteFile(cfg, []byte(`openai:
  api_key: dummy
`), 0o600)
	require.NoError(t, err)

	return cfg
}

type stubAuth struct{}

func (stubAuth) SaveToken(string) error    { return nil }
func (stubAuth) GetToken() (string, error) { return "token", nil }

type stubConfig struct{}

func (stubConfig) WriteField(string, any) error { return nil }
func (stubConfig) ReadField(string) any         { return nil }

type stubTokenFetcher struct{}

func (stubTokenFetcher) Fetch() (string, error) { return "token", nil }

type stubTokenStorer struct{}

func (stubTokenStorer) Store(string) error { return nil }

type stubEncryptor struct{}

func (stubEncryptor) Encrypt(v []byte) ([]byte, error) { return v, nil }

type stubDecryptor struct{}

func (stubDecryptor) Decrypt(v []byte) ([]byte, error) { return v, nil }
