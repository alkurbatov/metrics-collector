package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alkurbatov/metrics-collector/internal/app"
)

func main() {
	app := app.NewServer()

	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	ctx, cancel := context.WithCancel(context.Background())
	go app.Serve(ctx)

	signal := <-sigChan
	cancel()
	app.Shutdown(signal)
}
