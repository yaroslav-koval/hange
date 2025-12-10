package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const flagNameOmitContext = "omit-context"

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:     "explain",
	Short:   "Explain file(s) or directory(ies)",
	Long:    `Explain file(s) or directory(ies) from the engineer's perspective`,
	Example: `hange explain file1 file2 directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO

		// option 1
		// read provided files/dirs
		// read current dir (if flag not provided)
		// compress them

		ok, err := cmd.LocalFlags().GetBool(flagNameOmitContext)
		if err != nil {
			return err
		}

		fmt.Println("Flag received:", ok)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)

	explainCmd.Flags().BoolP(flagNameOmitContext, "o", false,
		"don't use current dir in context (worse reasoning, better security)")
}
