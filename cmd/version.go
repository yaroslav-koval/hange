package cmd

import (
	"bytes"
	"fmt"
	"os"

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

		cfg := make([]byte, len(buildConfig))
		copy(cfg, buildConfig)

		if err := r.ReadConfig(bytes.NewBuffer(cfg)); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to read version: %s\n", err)
		}

		_, err := fmt.Fprintln(os.Stderr, fmt.Sprintf("hange version %s", r.GetString("version")))
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
