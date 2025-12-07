package configcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

var configName = "." + consts.AppName

func TestInitCLIConfig(t *testing.T) {
	th := &tempHome{}

	t.Run("test create if not exists", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)

		// config doesn't exist at this moment, InitCLIConfig should initialize a new config

		err := initCLIConfig(viper.New(), "")
		require.NoError(t, err)

		dir, err := os.ReadDir(tempHomeDir)
		require.NoError(t, err)

		var found bool
		for _, entry := range dir {
			if entry.Name() == configName {
				found = true
				break
			}
		}

		assert.Truef(t, found, "config should be created in $HOME directory if config file is not provided")
	})

	t.Run("test keep values if exists", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)

		cfgPath := tempHomeDir + "/" + configName

		configContent := []byte("version: v1")

		err := os.WriteFile(cfgPath, configContent, 0644)
		require.NoError(t, err)

		err = initCLIConfig(viper.New(), "")
		require.NoError(t, err)

		actualFileContent, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Equalf(t, configContent, actualFileContent,
			"config shouldn't be changed if already existed in $HOME directory",
		)
	})

	t.Run("test custom config exists", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, configName)

		configContent := []byte("version: v1")

		err := os.WriteFile(cfgPath, configContent, 0644)
		require.NoError(t, err)

		err = initCLIConfig(viper.New(), cfgPath)
		require.NoError(t, err)

		actualFileContent, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Equalf(t, configContent, actualFileContent,
			"config shouldn't be changed",
		)
	})

	t.Run("test custom config not exists", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, configName)

		err := initCLIConfig(viper.New(), cfgPath)
		require.ErrorIsf(t, err, os.ErrNotExist, "config shouldn't be created by custom path if it doesn't exist")
	})

	t.Run("read value from config", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := filepath.Join(tempDir, configName)

		configContent := []byte("version: v1")

		err := os.WriteFile(cfgPath, configContent, 0644)
		require.NoError(t, err)

		vip := viper.New()

		err = initCLIConfig(vip, cfgPath)
		require.NoError(t, err)

		val := vip.GetString("version")
		assert.Equal(t, "v1", val)
	})

	t.Run("fails to create default config when home is invalid", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)

		invalidHome := filepath.Join(tempHomeDir, "missing")
		require.NoError(t, os.Setenv(homeEnvKey, invalidHome))

		err := initCLIConfig(viper.New(), "")
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
}

func TestNewCLIConfig(t *testing.T) {
	th := &tempHome{}

	t.Run("creates default config in home", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)

		cfg, err := NewCLIConfig("")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		_, err = os.Stat(filepath.Join(tempHomeDir, configName))
		require.NoError(t, err)
	})

	t.Run("fails when custom config missing", func(t *testing.T) {
		tempHomeDir := th.createTempHome(t)
		defer th.restoreHome(t)
		_ = tempHomeDir

		cfgPath := filepath.Join(t.TempDir(), configName)

		_, err := NewCLIConfig(cfgPath)
		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
	})
}

// WARN do not use in parallel
type tempHome struct {
	originalHome string
}

const homeEnvKey = "HOME"

func (th *tempHome) createTempHome(t *testing.T) string {
	t.Helper()

	th.originalHome = os.Getenv(homeEnvKey)

	tempHomeDir := t.TempDir()
	err := os.Setenv(homeEnvKey, tempHomeDir)
	require.NoError(t, err)

	return tempHomeDir
}

func (th *tempHome) restoreHome(t *testing.T) {
	t.Helper()

	err := os.Setenv(homeEnvKey, th.originalHome)
	require.NoError(t, err)
}
