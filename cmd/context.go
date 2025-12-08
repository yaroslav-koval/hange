package cmd

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

type appKey struct{}

var appContextKey appKey

var errAppNotInitialized = errors.New("application not initialized")

func appFromCmdContext(cmd *cobra.Command) *factory.App {
	v := cmd.Context().Value(appContextKey)

	app, ok := v.(*factory.App)

	if !ok {
		cobra.CheckErr(errAppNotInitialized)
	}

	return app
}

func appToCmdContext(cmd *cobra.Command, app *factory.App) context.Context {
	return context.WithValue(cmd.Context(), appContextKey, app)
}
