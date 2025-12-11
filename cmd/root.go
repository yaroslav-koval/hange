package cmd

import (
	"context"
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/envs"
	"github.com/yaroslav-koval/hange/pkg/factory"
	"github.com/yaroslav-koval/hange/pkg/factory/appfactory"
)

var (
	cfgPath    = os.Getenv(envs.EnvHangeConfigPath)
	cancel     context.CancelFunc
	appFactory = appfactory.NewCLIFactory
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hange",
	Short: "A reliable CLI soldier to perform routine tasks",
	Long: `A reliable CLI soldier to perform developer's routine tasks. 
It likes to explain code, write documentation and just chat.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		ctx, c := context.WithCancel(cmd.Context())
		cancel = c
		cmd.SetContext(ctx)

		sigCh := makeSignalChan()

		go func() {
			slog.Debug("Listening for a termination signal")
			<-sigCh
			slog.Info("Terminated")
			cancel()
		}()

		f := appFactory(cfgPath)
		app, err := factory.BuildApp(f)
		if err != nil {
			return err
		}

		ctx = appToCmdContext(cmd, &app)
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		cancel()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file (default is $HOME/.hange.yaml)")
}

func makeSignalChan() <-chan os.Signal {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	return sigs
}
