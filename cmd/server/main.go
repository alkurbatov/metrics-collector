package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/server"
	"github.com/rs/zerolog/log"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	cfg, err := config.NewServer()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	logging.Setup(cfg.Debug)
	log.Info().Msg(cfg.String())

	app := server.New(cfg)

	log.Info().Msg("Build version: " + buildVersion)
	log.Info().Msg("Build date: " + buildDate)
	log.Info().Msg("Build commit: " + buildCommit)

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
