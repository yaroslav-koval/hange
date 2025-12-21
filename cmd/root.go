package cmd

import (
	_ "embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/envs"
	"github.com/yaroslav-koval/hange/pkg/factory"
	"github.com/yaroslav-koval/hange/pkg/factory/agentfactory"
	"github.com/yaroslav-koval/hange/pkg/factory/appfactory"
)

const (
	flagKeyVerbose    = "verbose"
	flagKeyConfigPath = "config"
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

		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-ctx.Done()
			slog.Info("Terminated")
			cancel()
		}()

		cmd.SetContext(ctx)

		cfgPath, err := cmd.Flags().GetString(flagKeyConfigPath)
		if err != nil {
			return err
		}

		cliFactory := appfactory.NewCLIFactory(cfgPath)
		openAIFactory := agentfactory.NewOpenAIFactory()

		app := factory.NewAppBuilder(cliFactory, openAIFactory)

		ctx = appToContext(cmd.Context(), app)
		cmd.SetContext(ctx)

		return nil
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

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.PersistentFlags().String(flagKeyConfigPath, os.Getenv(envs.EnvHangeConfigPath), "config file (default is $HOME/.hange.yaml)")
	rootCmd.PersistentFlags().BoolP(flagKeyVerbose, "v", false, "verbose logging")
}
