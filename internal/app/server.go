package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/handlers"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/prof"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	flag "github.com/spf13/pflag"
)

type ServerConfig struct {
	ListenAddress  entity.NetAddress    `env:"ADDRESS"`
	StoreInterval  time.Duration        `env:"STORE_INTERVAL"`
	StorePath      string               `env:"STORE_FILE"`
	RestoreOnStart bool                 `env:"RESTORE"`
	Secret         security.Secret      `env:"KEY"`
	DatabaseURL    security.DatabaseURL `env:"DATABASE_DSN"`
	PprofAddress   entity.NetAddress    `env:"PPROF_ADDRESS"`
	Debug          bool                 `env:"DEBUG"`
}

func NewServerConfig() (*ServerConfig, error) {
	var (
		listenAddress entity.NetAddress = "0.0.0.0:8080"
		pprofAddress  entity.NetAddress
	)

	flag.VarP(
		&listenAddress,
		"listen-address",
		"a",
		"address:port server listens on",
	)

	storeInterval := flag.DurationP(
		"store-interval",
		"i",
		300*time.Second,
		"count of seconds after which metrics are dumped to the disk, zero value activates saving after each request",
	)
	storePath := flag.StringP(
		"store-file",
		"f",
		"/tmp/devops-metrics-db.json",
		"path to file to store metrics",
	)
	restoreOnStart := flag.BoolP(
		"restore",
		"r",
		true,
		"whether to restore state on startup or not",
	)
	secret := security.Secret("")
	flag.VarP(
		&secret,
		"key",
		"k",
		"secret key for signature generation",
	)

	databaseURL := flag.StringP(
		"db-dsn",
		"d",
		"",
		"full database connection URL",
	)

	flag.VarP(
		&pprofAddress,
		"pprof-address",
		"p",
		"enable pprof on specified address:port",
	)

	debug := flag.BoolP(
		"debug",
		"g",
		false,
		"enable verbose logging",
	)

	flag.Parse()

	cfg := &ServerConfig{
		ListenAddress:  listenAddress,
		StorePath:      *storePath,
		StoreInterval:  *storeInterval,
		RestoreOnStart: *restoreOnStart,
		Secret:         secret,
		DatabaseURL:    security.DatabaseURL(*databaseURL),
		Debug:          *debug,
		PprofAddress:   pprofAddress,
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}

	if len(cfg.StorePath) == 0 && cfg.RestoreOnStart {
		return nil, entity.ErrRestoreNoSource
	}

	return cfg, nil
}

func (c ServerConfig) String() string {
	var sb strings.Builder

	sb.WriteString("Configuration:\n")
	sb.WriteString(fmt.Sprintf("\t\tListening address: %s\n", c.ListenAddress))

	sb.WriteString(fmt.Sprintf("\t\tStore interval: %s\n", c.StoreInterval))
	sb.WriteString(fmt.Sprintf("\t\tStore path: %s\n", c.StorePath))
	sb.WriteString(fmt.Sprintf("\t\tRestore on start: %t\n", c.RestoreOnStart))

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
	}

	if len(c.DatabaseURL) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tDatabase URL: %s\n", c.DatabaseURL))
	}

	if len(c.PprofAddress) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tPprof address: %s\n", c.PprofAddress))
	}

	sb.WriteString(fmt.Sprintf("\t\tDebug: %t\n", c.Debug))

	return sb.String()
}

type Server struct {
	Config     *ServerConfig
	Storage    storage.Storage
	HTTPServer *http.Server
	Profiler   *prof.Profiler
}

func NewServer() *Server {
	cfg, err := NewServerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	logging.Setup(cfg.Debug)
	log.Info().Msg(cfg.String())

	var pool *pgxpool.Pool

	if len(cfg.DatabaseURL) > 0 {
		if err = runMigrations(cfg.DatabaseURL); err != nil {
			log.Fatal().Err(err).Msg("")
		}

		pool, err = pgxpool.New(context.Background(), string(cfg.DatabaseURL))
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}

	dataStore := storage.NewDataStore(pool, cfg.StorePath, cfg.StoreInterval)
	recorder := services.NewMetricsRecorder(dataStore)
	healthcheck := services.NewHealthCheck(dataStore)

	var signer *security.Signer
	if len(cfg.Secret) > 0 {
		signer = security.NewSigner(cfg.Secret)
	}

	router := handlers.Router(cfg.ListenAddress, "./web/views", recorder, healthcheck, signer)
	srv := &http.Server{
		Addr:    cfg.ListenAddress.String(),
		Handler: router,

		// NB (alkurbatov): Set reasonable timeouts, see:
		// https://habr.com/ru/company/ispring/blog/560032/
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	var profiler *prof.Profiler
	if len(cfg.PprofAddress) > 0 {
		profiler = prof.New(cfg.PprofAddress)
	}

	return &Server{
		Config:     cfg,
		Storage:    dataStore,
		HTTPServer: srv,
		Profiler:   profiler,
	}
}

func (app *Server) restoreStorage() {
	fileStore, ok := app.Storage.(*storage.FileBackedStorage)
	if !ok {
		log.Warn().Msg("Metrics storage backend doesn't support restoring from disk!")
		return
	}

	if err := fileStore.Restore(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func (app *Server) dumpStorage(ctx context.Context) {
	if _, ok := app.Storage.(*storage.FileBackedStorage); !ok {
		log.Warn().Msg("Metrics storage backend doesn't support saving to disk!")
		return
	}

	ticker := time.NewTicker(app.Config.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer tryRecover()

				if err := app.Storage.(*storage.FileBackedStorage).Dump(ctx); err != nil {
					log.Error().Err(err).Msg("")
				}
			}()

		case <-ctx.Done():
			log.Info().Msg("Shutdown storage dumping")
			return
		}
	}
}

func (app *Server) Serve(ctx context.Context) {
	if len(app.Config.DatabaseURL) == 0 && len(app.Config.StorePath) > 0 {
		if app.Config.RestoreOnStart {
			app.restoreStorage()
		}

		if app.Config.StoreInterval > 0 {
			go app.dumpStorage(ctx)
		}
	}

	if app.Profiler != nil {
		go app.Profiler.Start()
	}

	if err := app.HTTPServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("")
	}
}

func (app *Server) Shutdown(signal os.Signal) {
	log.Info().Msg(fmt.Sprintf("Signal '%s' received, shutting down...", signal))

	if err := app.HTTPServer.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("")
	}

	if err := app.Storage.Close(); err != nil {
		log.Error().Err(err).Msg("")
	}

	if app.Profiler != nil {
		if err := app.Profiler.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	log.Info().Msg("Successfully shutdown the service")
}
