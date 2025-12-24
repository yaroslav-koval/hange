package cmd

import (
	"context"
	"errors"

	"github.com/yaroslav-koval/hange/domain/factory"
)

type appKey struct{}

var appContextKey appKey

var errAppNotInitialized = errors.New("application not initialized")

func appFromContext(ctx context.Context) (factory.AppBuilder, error) {
	v := ctx.Value(appContextKey)

	app, ok := v.(factory.AppBuilder)
	if !ok {
		return nil, errAppNotInitialized
	}

	return app, nil
}

func appToContext(ctx context.Context, app factory.AppBuilder) context.Context {
	return context.WithValue(ctx, appContextKey, app)
}
