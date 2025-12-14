package cmd

import (
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:     "commit [input]",
	Short:   "Makes a commit to current git branch",
	Long:    `Takes a changelist of git, generates a commit message and commits to a current branch.`,
	Example: `hange commit "Task description"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := appFromContext(cmd.Context())
		if err != nil {
			return err
		}

		message, err := generateCommitMessage(cmd.Context(), app, args)
		if err != nil {
			return err
		}

		return app.Git.Commit(cmd.Context(), message)
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
