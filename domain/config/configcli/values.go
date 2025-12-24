package configcli

func (c *viperConfigurator) WriteField(field string, value any) error {
	c.viper.Set(field, value)

	err := c.viper.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}

func (c *viperConfigurator) ReadField(field string) any {
	// if Viper's AutomaticEnv is enabled, it tries to read value not only from config, but also from environment variables
	return c.viper.Get(field)
}
