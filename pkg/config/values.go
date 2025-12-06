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
	return viper.Get(field)
}
