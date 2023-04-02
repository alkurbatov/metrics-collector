package main

import (
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
	cfg, err := config.NewAgent()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	logging.Setup(cfg.Debug)
	log.Info().Msg(cfg.String())

	app, err := agent.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	log.Info().Msg("Build version: " + buildVersion)
	log.Info().Msg("Build date: " + buildDate)
	log.Info().Msg("Build commit: " + buildCommit)

	app.Run()
}
