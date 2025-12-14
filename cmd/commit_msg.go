package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/agent/entity"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

var commitMsgCmd = &cobra.Command{
	Use:     "commit-msg [input]",
	Short:   "Generate a git commit message",
	Long:    `Takes a changelist of git and outputs a short commit message.`,
	Example: `hange commit-msg "Task description"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := appFromContext(cmd.Context())
		if err != nil {
			return err
		}

		message, err := generateCommitMessage(cmd.Context(), app, args)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), message)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitMsgCmd)
}

func generateCommitMessage(ctx context.Context, app *factory.App, args []string) (string, error) {
	if len(args) > 1 {
		return "", fmt.Errorf("received %d args. This command accepts at most 1 arg with user context of changes",
			len(args))
	}

	var userInput string
	if len(args) == 1 {
		userInput = args[0]
	}

	status, err := app.Git.Status(ctx)
	if err != nil {
		return "", err
	}

	stagedStatus, err := app.Git.StagedStatus(ctx)
	if err != nil {
		return "", err
	}

	diff, err := app.Git.StagedDiff(ctx, 30)
	if err != nil {
		return "", err
	}

	res, err := app.Agent.CreateCommitMessage(ctx, entity.CommitData{
		UserInput:    userInput,
		Status:       status,
		StagedStatus: stagedStatus,
		Diff:         diff,
	})
	if err != nil {
		return "", err
	}

	return res, nil
}
