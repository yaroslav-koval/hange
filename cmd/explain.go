package cmd

import (
	"github.com/spf13/cobra"
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain file(s) or directory(ies)",
	Long:  `Explain file(s) or directory(ies) from the engineer's perspective`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO

		return nil
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
}
