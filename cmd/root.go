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

const (
	flagKeyVerbose    = "verbose"
	flagKeyConfigPath = "config"
)

var (
	appBuilder factory.AppBuilder
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hange",
	Short: "A reliable CLI soldier to perform routine tasks",
	Long: `A reliable CLI soldier to perform developer's routine tasks.
It likes to explain code, write documentation and just chat.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		isVerbose, err := cmd.Flags().GetBool(flagKeyVerbose)
		if err != nil {
			return err
		}

		if isVerbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		ctx, cancel := context.WithCancel(cmd.Context())
		cmd.SetContext(ctx)

		sigCh := makeOsSignalChan()

		go func() {
			<-sigCh
			slog.Info("Terminated")
			cancel()
		}()

		cfgPath, err := cmd.Flags().GetString(flagKeyConfigPath)
		if err != nil {
			return err
		}

		app, err := appBuilder.BuildApp(appfactory.NewCLIFactory(cfgPath))
		if err != nil {
			return err
		}

		ctx = appToCmdContext(cmd, app)
		cmd.SetContext(ctx)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	factory.NewAppBuilder()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP(flagKeyConfigPath, "", os.Getenv(envs.EnvHangeConfigPath), "config file (default is $HOME/.hange.yaml)")
	rootCmd.PersistentFlags().BoolP(flagKeyVerbose, "v", false, "verbose logging")
}

func makeOsSignalChan() <-chan os.Signal {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	return sigs
}
