package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/agent"
	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/rs/zerolog/log"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	cfg := config.NewAgent()

	logging.Setup(cfg.Debug)

	log.Info().Msg("Build version: " + buildVersion)
	log.Info().Msg("Build date: " + buildDate)
	log.Info().Msg("Build commit: " + buildCommit)
	log.Info().Msg(cfg.String())

	app := agent.New(cfg)

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
