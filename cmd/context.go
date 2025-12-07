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

func appFromCtx(cmd *cobra.Command) *factory.App {
	return cmd.Context().Value(appContextKey).(*factory.App)
}

func appToCtx(cmd *cobra.Command, app *factory.App) context.Context {
	return context.WithValue(cmd.Context(), appContextKey, app)
}
