// Package server implements metrics collecting service.
package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/config"
	"github.com/alkurbatov/metrics-collector/internal/handlers"
	"github.com/alkurbatov/metrics-collector/internal/prof"
	"github.com/alkurbatov/metrics-collector/internal/recovery"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Server struct {
	config   *config.Server
	storage  storage.Storage
	server   *http.Server
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

func (app *Server) Serve(ctx context.Context) {
	if len(app.config.DatabaseURL) == 0 && len(app.config.StorePath) > 0 {
		if app.config.RestoreOnStart {
			app.restoreStorage()
		}

		if app.config.StoreInterval > 0 {
			go app.dumpStorage(ctx)
		}
	}

	if app.profiler != nil {
		go app.profiler.Start()
	}

	if err := app.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("")
	}
}

func (app *Server) Shutdown(signal os.Signal) {
	log.Info().Msg(fmt.Sprintf("Signal '%s' received, shutting down...", signal))

	if err := app.server.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("")
	}

	if err := app.storage.Close(); err != nil {
		log.Error().Err(err).Msg("")
	}

	if app.profiler != nil {
		if err := app.profiler.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("")
		}
	}

	log.Info().Msg("Successfully shutdown the service")
}
