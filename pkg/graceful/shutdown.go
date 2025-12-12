package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func Shutdown(ctx context.Context) context.Context {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-sigs
		cancel()
	}()

	return ctx
}
