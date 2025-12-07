package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

const configTypeYaml = "yaml"

type Configurator interface {
	WriteField(field string, value any) error
	ReadField(field string) any
}

func NewCLIConfig(cfgFile string) (Configurator, error) {
	conf := &viperConfigurator{
		viper: viper.New(),
	}

	if err := initCLIConfig(conf.viper, cfgFile); err != nil {
		return nil, err
	}

	return conf, nil
}

type viperConfigurator struct {
	viper *viper.Viper
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
