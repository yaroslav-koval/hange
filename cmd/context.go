package cmd

import (
	"context"
	"errors"

	"github.com/yaroslav-koval/hange/pkg/factory"
)

type appKey struct{}

var appContextKey appKey

var errAppNotInitialized = errors.New("application not initialized")

func appFromContext(ctx context.Context) (*factory.App, error) {
	v := ctx.Value(appContextKey)

	app, ok := v.(*factory.App)
	if !ok {
		return nil, errAppNotInitialized
	}

	return app, nil
}

func appToContext(ctx context.Context, app *factory.App) context.Context {
	return context.WithValue(ctx, appContextKey, app)
}
