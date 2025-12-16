package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/factory"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
	"golang.org/x/sync/errgroup"
)

var explainCmd = &cobra.Command{
	Use:     "explain [inputs]",
	Short:   "Explain file(s) or directory(ies)",
	Long:    `Explain file(s) or directory(ies) from the engineer's perspective.`,
	Example: `hange explain file1 file2 directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := appFromContext(cmd.Context())
		if err != nil {
			return err
		}

		ep := &explainCmdProcessor{
			app: app,
		}

		if err := ep.validateArgs(args); err != nil {
			return err
		}

		e, err := ep.processExplanation(cmd.Context(), args)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		}

		if _, err = fmt.Fprintln(cmd.OutOrStdout(), e); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
}

type explainCmdProcessor struct {
	app *factory.App
}

var errNoArgs = errors.New("no arguments provided")
var errEmptyArg = errors.New("empty argument")

func (ep *explainCmdProcessor) validateArgs(args []string) error {
	if len(args) == 0 {
		return errNoArgs
	}

	for _, arg := range args {
		if len(arg) == 0 {
			return errEmptyArg
		}
	}

	return nil
}

func (ep *explainCmdProcessor) processExplanation(ctx context.Context, args []string) (string, error) {
	eg, ctx := errgroup.WithContext(ctx)

	fileNames, err := ep.app.FileProvider.GetAllFileNames(ctx, args)
	if err != nil {
		return "", err
	}

	filesCh, doneCh := ep.app.FileProvider.ReadFiles(ctx, fileprovider.Config{
		Workers:    3, // TODO take from config
		BufferSize: 3 * 2,
	}, fileNames)

	eg.Go(func() error {
		return <-doneCh
	})

	var output string

	eg.Go(func() error {
		output, err = ep.app.Agent.ExplainFiles(ctx, filesCh)
		if err != nil {
			return err
		}

		return nil
	})

	if err = eg.Wait(); err != nil {
		return "", err
	}

	return output, nil
}
