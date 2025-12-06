package config

import (
	"github.com/spf13/viper"
)

func WriteField(field string, value any) error {
	viper.Set(field, value)

	err := viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}

func ReadField(field string) any {
	// if Viper's AutomaticEnv is enabled, it tries to read value not only from config, but also from environment variables
	return viper.Get(field)
}
