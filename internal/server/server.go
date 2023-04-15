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
	"github.com/alkurbatov/metrics-collector/internal/grpcbackend"
	"github.com/alkurbatov/metrics-collector/internal/grpcserver"
	"github.com/alkurbatov/metrics-collector/internal/httpbackend"
	"github.com/alkurbatov/metrics-collector/internal/httpserver"
	"github.com/alkurbatov/metrics-collector/internal/prof"
	"github.com/alkurbatov/metrics-collector/internal/recovery"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const _defaultShutdownTimeout = 60 * time.Second

type Server struct {
	// Full configuration of the service.
	config *config.Server

	// Storage backend (in memory, file, database).
	storage storage.Storage

	// Instance of HTTP server providing HTTP API.
	httpServer *httpserver.Server

	// Instance of gRPC server providing gRPC API.
	grpcServer *grpcserver.Server

	// Instance of HTTP server serving pprof endpoints.
	// Works on different port.
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

	var key security.PrivateKey
	if len(cfg.PrivateKeyPath) != 0 {
		key, err = security.NewPrivateKey(cfg.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("server - New - security.NewPrivateKey: %w", err)
		}
	}

	view, err := template.ParseFiles("./web/views/metrics.html")
	if err != nil {
		return nil, fmt.Errorf("Server - New - template.ParseFiles: %w", err)
	}

	router := httpbackend.Router(cfg.Address, view, recorder, healthcheck, signer, key, cfg.TrustedSubnet)
	httpSrv := httpserver.New(router, cfg.Address)

	grpcSrv := grpcserver.New(cfg.GRPCAddress)
	grpcbackend.NewHealthServer(grpcSrv.Instance(), healthcheck)

	profiler := prof.New(cfg.PprofAddress)

	return &Server{
		config:     cfg,
		storage:    dataStore,
		httpServer: httpSrv,
		grpcServer: grpcSrv,
		profiler:   profiler,
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

// Run starts the main app and waits till compeletion or termination signal.
func (app *Server) Run() {
	ctx, cancelBackgroundTasks := context.WithCancel(context.Background())

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	if len(app.config.DatabaseURL) == 0 && len(app.config.StorePath) > 0 {
		if app.config.RestoreOnStart {
			app.restoreStorage()
		}

		if app.config.StoreInterval > 0 {
			go app.dumpStorage(ctx)
		}
	}

	app.httpServer.Start()
	app.grpcServer.Start()

	select {
	case s := <-interrupt:
		log.Info().Msg("app - Run - interrupt: signal " + s.String())
	case err := <-app.httpServer.Notify():
		log.Error().Err(err).Msg("app - Run - app.httpServer.Notify")
	case err := <-app.grpcServer.Notify():
		log.Error().Err(err).Msg("app - Run - app.grpcServer.Notify")
	case err := <-app.profiler.Notify():
		log.Error().Err(err).Msg("app - Run - app.profiler.Notify")
	}

	log.Info().Msg("Shutting down...")

	stopped := make(chan struct{})

	stopCtx, cancel := context.WithTimeout(context.Background(), _defaultShutdownTimeout)
	defer cancel()

	cancelBackgroundTasks()

	go func() {
		app.shutdown(stopCtx)
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info().Msg("Server shutdown successful")

	case <-stopCtx.Done():
		log.Warn().Msgf("Exceeded %s shutdown timeout, exit forcibly", _defaultShutdownTimeout)
	}
}

func (app *Server) shutdown(ctx context.Context) {
	log.Info().Msg("Shutting down HTTP API...")

	if err := app.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("")
	}

	log.Info().Msg("Shutting down gRPC API...")
	app.grpcServer.Shutdown()

	log.Info().Msg("Shutting down storage backend...")

	if err := app.storage.Close(ctx); err != nil {
		log.Error().Err(err).Msg("")
	}

	if err := app.profiler.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("")
	}
}
