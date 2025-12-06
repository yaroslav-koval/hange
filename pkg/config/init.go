package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

const configTypeYaml = "yaml"

func InitCLIConfig(cfgFile string) error {
	if err := setConfigFileOrDefault(cfgFile); err != nil {
		return err
	}

	viper.SetEnvPrefix(consts.AppName)
	viper.AutomaticEnv()

	slog.Info("Using config file: " + viper.ConfigFileUsed())

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func setConfigFileOrDefault(cfgFile string) error {
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		cfgFile = fmt.Sprintf("%s/.%s", home, consts.AppName)

		// create a config file if not exists
		if _, err = os.OpenFile(cfgFile, os.O_CREATE, 0600); err != nil {
			return err
		}
	}

	viper.SetConfigType(configTypeYaml)
	viper.SetConfigFile(cfgFile)

	return nil
}
