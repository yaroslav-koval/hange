package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

func TestWriteField(t *testing.T) {
	t.Parallel()

	t.Run("config not provided error", func(t *testing.T) {
		conf := &configurator{viper: viper.New()}
		err := conf.writeField("field", "value")
		assert.ErrorAs(t, err, &viper.ConfigFileNotFoundError{})
	})

	t.Run("success", func(t *testing.T) {
		tempDir := t.TempDir()
		cfgPath := tempDir + "/." + consts.AppName

		configContent := "version: v1"

		require.NoError(t, os.WriteFile(cfgPath, []byte(configContent), 0600))

		conf := &configurator{viper: viper.New()}
		require.NoError(t, setUpViperConfig(conf.viper, cfgPath))

		require.NoError(t, conf.writeField("token", "secret-value"))

		actualFile, err := os.ReadFile(cfgPath)
		require.NoError(t, err)
		assert.Contains(t, string(actualFile), configContent)
		assert.Contains(t, string(actualFile), "token: secret-value")
	})
}

func TestReadField(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cfgPath := tempDir + "/." + consts.AppName

	configContent := "version: v1"

	require.NoError(t, os.WriteFile(cfgPath, []byte(configContent), 0600))

	conf := &configurator{viper: viper.New()}
	require.NoError(t, setUpViperConfig(conf.viper, cfgPath))

	actualValue := conf.readField("version")
	assert.Equal(t, "v1", actualValue)
}

func setUpViperConfig(viper *viper.Viper, cfgPath string) error {
	viper.SetConfigFile(cfgPath)
	viper.SetConfigType(configTypeYaml)
	return viper.ReadInConfig()
}
