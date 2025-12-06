package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

const configTypeYaml = "yaml"

type configurator struct {
	viper *viper.Viper
}

var globalConfigurator = &configurator{}

func InitCLIConfig(cfgFile string) error {
	globalConfigurator.viper = viper.New()

	return initCLIConfig(globalConfigurator.viper, cfgFile)
}

func initCLIConfig(viper *viper.Viper, cfgFile string) error {
	if err := setConfigFileOrDefault(viper, cfgFile); err != nil {
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

func setConfigFileOrDefault(viper *viper.Viper, cfgFile string) error {
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
