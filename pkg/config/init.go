package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yaroslav-koval/hange/pkg/consts"
)

// InitCLIConfig reads in config file and ENV variables if set.
func InitCLIConfig(cfgFile string) error {
	err := setConfigFileOrDefault(cfgFile)
	cobra.CheckErr(err)

	viper.SetEnvPrefix(consts.AppName)
	viper.AutomaticEnv()

	slog.Info("Using config file: " + viper.ConfigFileUsed())

	if err = viper.ReadInConfig(); err != nil {
		if errors.Is(err, viper.ConfigFileNotFoundError{}) {
			return viper.WriteConfig()
		}

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
		if _, err = os.OpenFile(cfgFile, os.O_CREATE, 0644); err != nil {
			return err
		}
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgFile)

	return nil
}
