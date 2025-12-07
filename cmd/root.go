package cmd

import (
	_ "embed"
	"os"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/envs"
	"github.com/yaroslav-koval/hange/pkg/factory"
	"github.com/yaroslav-koval/hange/pkg/factory/factorycli"
)

var cfgFile = os.Getenv(envs.EnvHangeConfigPath)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hange",
	Short: "A reliable CLI soldier to perform routine tasks",
	Long: `A reliable CLI soldier to perform developer's routine tasks. 
It likes to explain code, write documentation and just chat.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f := factorycli.NewCLIFactory(cfgFile)

		app, err := factory.BuildApp(f)
		if err != nil {
			return err
		}

		ctx := appToCtx(cmd, &app)
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

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.hange.yaml)")
}
