package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/yaroslav-koval/hange/pkg/factory"
)

func TestRootCommandConfiguration(t *testing.T) {
	require.Equal(t, "hange", rootCmd.Use)

	flag := rootCmd.PersistentFlags().Lookup("config")
	require.NotNil(t, flag)
	require.Equal(t, "config", flag.Name)
}

func TestPersistentPreRunESetsAppContext(t *testing.T) {
	cfg := writeTempConfig(t)

	originalCfgFile := cfgFile
	t.Cleanup(func() { cfgFile = originalCfgFile })
	cfgFile = cfg

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

	originalCfgFile := cfgFile
	t.Cleanup(func() { cfgFile = originalCfgFile })
	cfgFile = missingCfg

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.Error(t, err)
}

func TestExecuteSucceeds(t *testing.T) {
	cfg := writeTempConfig(t)

	originalCfgFile := cfgFile
	t.Cleanup(func() {
		cfgFile = originalCfgFile
		buildConfig = nil
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(context.Background())
	})
	cfgFile = cfg

	SetBuildConfig([]byte("version: 1.2.3"))

	rootCmd.SetArgs([]string{"--config", cfg, "version"})
	rootCmd.SetContext(context.Background())

	Execute() // should not call os.Exit on success
}

func writeTempConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	cfg := filepath.Join(dir, ".hange.yaml")
	err := os.WriteFile(cfg, []byte("openai:\n  api_key: dummy\n"), 0o600)
	require.NoError(t, err)

	return cfg
}
