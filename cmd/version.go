package cmd

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Hange version",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := viper.New()
		r.SetConfigType("yaml")

		buildCfg := getBuildConfig()

		if err := r.ReadConfig(bytes.NewBuffer(buildCfg)); err != nil {
			slog.Info(fmt.Sprintf("Failed to read version: %s\n", err))
		}

		slog.Info(fmt.Sprintf("hange version %s", r.GetString("version")))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
