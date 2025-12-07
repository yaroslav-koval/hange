package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/config"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the Hange version",
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := config.ReadFieldFromBytes(getBuildConfig(), config.FileTypeYaml, "version")
		if err != nil {
			slog.Info(fmt.Sprintf("Failed to read version: %s\n", err))
			return err
		}

		if val == nil {
			slog.Info(fmt.Sprintf("Failed to read version: version value is nil"))
			return err
		}

		slog.Info(fmt.Sprintf("hange version %v", val))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
