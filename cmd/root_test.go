package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestRootCommandFlags(t *testing.T) {
	require.Equal(t, "hange", rootCmd.Use)

	flags := rootCmd.PersistentFlags()

	configFlag := flags.Lookup(flagKeyConfigPath)
	require.NotNil(t, configFlag)
	require.Equal(t, "", configFlag.DefValue)

	verboseFlag := flags.Lookup(flagKeyVerbose)
	require.NotNil(t, verboseFlag)
	require.Equal(t, "v", verboseFlag.Shorthand)
	require.Equal(t, "false", verboseFlag.DefValue)

	originalCfg, err := flags.GetString(flagKeyConfigPath)
	require.NoError(t, err)

	originalVerbose, err := flags.GetBool(flagKeyVerbose)
	require.NoError(t, err)

	configFlagChanged := configFlag.Changed
	verboseFlagChanged := verboseFlag.Changed

	t.Cleanup(func() {
		_ = flags.Set(flagKeyConfigPath, originalCfg)
		configFlag.Changed = configFlagChanged

		_ = flags.Set(flagKeyVerbose, strconv.FormatBool(originalVerbose))
		verboseFlag.Changed = verboseFlagChanged
	})

	newCfg := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, flags.Set(flagKeyConfigPath, newCfg))
	cfgValue, err := flags.GetString(flagKeyConfigPath)
	require.NoError(t, err)
	require.Equal(t, newCfg, cfgValue)

	require.NoError(t, flags.Set(flagKeyVerbose, "true"))
	verboseValue, err := flags.GetBool(flagKeyVerbose)
	require.NoError(t, err)
	require.True(t, verboseValue)
}

func TestRootPersistentPreRunEHandlesCancellation(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")

	cmd := &cobra.Command{}
	cmd.Flags().String(flagKeyConfigPath, cfgPath, "config")
	cmd.Flags().BoolP(flagKeyVerbose, "v", false, "verbose logging")
	cmd.SetContext(context.Background())

	err := rootCmd.PersistentPreRunE(cmd, nil)
	require.NoError(t, err)

	ctx := cmd.Context()

	proc, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, proc.Signal(unix.SIGTERM))

	require.Eventually(t, func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)

	require.ErrorIs(t, ctx.Err(), context.Canceled)

	app, err := appFromContext(ctx)
	require.NoError(t, err)
	require.NotNil(t, app)
}
