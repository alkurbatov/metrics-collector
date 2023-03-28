// Package server implements metrics collecting service.
package server

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/handlers"
	"github.com/alkurbatov/metrics-collector/internal/httpserver"
	"github.com/alkurbatov/metrics-collector/internal/prof"
	"github.com/alkurbatov/metrics-collector/internal/recovery"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const _defaultShutdownTimeout = 10 * time.Second

type Server struct {
	config   *config.Server
	storage  storage.Storage
	server   *httpserver.Server
	profiler *prof.Profiler
}

func New(cfg *config.Server) (*Server, error) {
	var (
		pool *pgxpool.Pool
		err  error
	)

	if len(cfg.DatabaseURL) > 0 {
		if err = runMigrations(cfg.DatabaseURL); err != nil {
			return nil, fmt.Errorf("Server - New - runMigrations: %w", err)
		}

		pool, err = pgxpool.New(context.Background(), string(cfg.DatabaseURL))
		if err != nil {
			return nil, fmt.Errorf("Server - New - pgxpool.New: %w", err)
		}
	}

	dataStore := storage.NewDataStore(pool, cfg.StorePath, cfg.StoreInterval)
	recorder := services.NewMetricsRecorder(dataStore)
	healthcheck := services.NewHealthCheck(dataStore)

	var signer *security.Signer
	if len(cfg.Secret) > 0 {
		signer = security.NewSigner(cfg.Secret)
	}

	view, err := template.ParseFiles("./web/views/metrics.html")
	if err != nil {
		return nil, fmt.Errorf("Server - New - template.ParseFiles: %w", err)
	}

	router := handlers.Router(cfg.ListenAddress, view, recorder, healthcheck, signer)
	srv := httpserver.New(router, cfg.ListenAddress)

	profiler := prof.New(cfg.PprofAddress)

	return &Server{
		config:   cfg,
		storage:  dataStore,
		server:   srv,
		profiler: profiler,
	}, nil
}

func (app *Server) restoreStorage() {
	fileStore, ok := app.storage.(*storage.FileBackedStorage)
	if !ok {
		log.Warn().Msg("Metrics storage backend doesn't support restoring from disk!")
		return
	}

	if err := fileStore.Restore(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func (app *Server) dumpStorage(ctx context.Context) {
	if _, ok := app.storage.(*storage.FileBackedStorage); !ok {
		log.Warn().Msg("Metrics storage backend doesn't support saving to disk!")
		return
	}

	ticker := time.NewTicker(app.config.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer recovery.TryRecover()

				if err := app.storage.(*storage.FileBackedStorage).Dump(ctx); err != nil {
					log.Error().Err(err).Msg("")
				}
			}()

		case <-ctx.Done():
			log.Info().Msg("Shutdown storage dumping")
			return
		}
	}
}

func (app *Server) Run() {
	ctx, cancelBackgroundTasks := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	if len(app.config.DatabaseURL) == 0 && len(app.config.StorePath) > 0 {
		if app.config.RestoreOnStart {
			app.restoreStorage()
		}

		if app.config.StoreInterval > 0 {
			go app.dumpStorage(ctx)
		}
	}

	select {
	case s := <-interrupt:
		log.Info().Msg("app - Run - interrupt: signal " + s.String())
	case err := <-app.server.Notify():
		log.Error().Err(err).Msg("app - Run - app.server.Notify")
	case err := <-app.profiler.Notify():
		log.Error().Err(err).Msg("app - Run - app.profiler.Notify")
	}

	log.Info().Msg("Shutting down...")

	stopped := make(chan struct{})

	stopCtx, cancel := context.WithTimeout(context.Background(), _defaultShutdownTimeout)
	defer cancel()

	cancelBackgroundTasks()

	go app.shutdown(stopCtx, stopped)

	select {
	case <-stopped:
		log.Info().Msg("Server shutdown successful")

	case <-stopCtx.Done():
		log.Warn().Msgf("Exceeded %s shutdown timeout, exit forcibly", _defaultShutdownTimeout)
	}
}

func (app *Server) shutdown(
	ctx context.Context,
	notify chan<- struct{},
) {
	if err := app.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("")
	}

	if err := app.storage.Close(ctx); err != nil {
		log.Error().Err(err).Msg("")
	}

	if app.profiler != nil {
		if err := app.profiler.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	close(notify)
}
