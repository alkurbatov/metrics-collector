package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/agent"
	"github.com/rs/zerolog/log"
)

func main() {
	app := agent.New()

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
	log.Info().Msg(fmt.Sprintf("Signal '%s' received, shutting down...", signal))

	cancel()

	// NB (alkurbatov): Give the goroutines some time to shutdown.
	time.Sleep(time.Second)
}
