package config

func WriteField(field string, value any) error {
	return globalConfigurator.writeField(field, value)
}

func (c *configurator) writeField(field string, value any) error {
	c.viper.Set(field, value)

	err := c.viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}

func ReadField(field string) any {
	// if Viper's AutomaticEnv is enabled, it tries to read value not only from config, but also from environment variables
	return globalConfigurator.readField(field)
}

func (c *configurator) readField(field string) any {
	// if Viper's AutomaticEnv is enabled, it tries to read value not only from config, but also from environment variables
	return c.viper.Get(field)
}
