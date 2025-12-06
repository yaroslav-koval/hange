package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

func TestInitCLIConfig(t *testing.T) {
	fileName := "." + consts.AppName
	th := &tempHome{}

	t.Run("test create if not exists", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)

		// config doesn't exist at this moment, InitCLIConfig should initialize a new config

		err := InitCLIConfig("")
		require.NoError(t, err)

		dir, err := os.ReadDir(tempHomeDir)
		require.NoError(t, err)

		var found bool
		for _, entry := range dir {
			if entry.Name() == fileName {
				found = true
				break
			}
		}

		assert.Truef(t, found, "config should be created in $HOME directory if config file is not provided")
	})

	t.Run("test keep values if exists", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)

		cfgPath := tempHomeDir + "/" + fileName

		configContent := []byte("version: v1")

		err := os.WriteFile(cfgPath, configContent, 0644)
		require.NoError(t, err)

		err = InitCLIConfig("")
		require.NoError(t, err)

		actualFileContent, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Equalf(t, configContent, actualFileContent,
			"config shouldn't be changed if already existed in $HOME directory",
		)
	})

	t.Run("test custom config exists", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := tempDir + "/" + fileName

		configContent := []byte("version: v1")

		err := os.WriteFile(cfgPath, configContent, 0644)
		require.NoError(t, err)

		err = InitCLIConfig(cfgPath)
		require.NoError(t, err)

		actualFileContent, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Equalf(t, configContent, actualFileContent,
			"config shouldn't be changed",
		)
	})

	t.Run("test custom config not exists", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := tempDir + "/" + fileName

		err := InitCLIConfig(cfgPath)
		require.ErrorIsf(t, err, os.ErrNotExist, "config shouldn't be created by custom path if it doesn't exist")
	})

	t.Run("read value from config", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := tempDir + "/" + fileName

		configContent := []byte("version: v1")

		err := os.WriteFile(cfgPath, configContent, 0644)
		require.NoError(t, err)

		err = InitCLIConfig(cfgPath)
		require.NoError(t, err)

		val := viper.GetString("version")
		assert.Equal(t, "v1", val)
	})
}

// WARN do not use in parallel
type tempHome struct {
	originalHome string
}

const homeEnvKey = "HOME"

func (th *tempHome) createTempHome(t *testing.T) string {
	th.originalHome = os.Getenv(homeEnvKey)

	tempHomeDir := t.TempDir()
	err := os.Setenv(homeEnvKey, tempHomeDir)
	require.NoError(t, err)

	return tempHomeDir
}

func (th *tempHome) restoreHome(t *testing.T) {
	err := os.Setenv(homeEnvKey, th.originalHome)
	require.NoError(t, err)
}
