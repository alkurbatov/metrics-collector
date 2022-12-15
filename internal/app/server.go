package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/handlers"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	ListenAddress  string        `env:"ADDRESS" envDefault:"0.0.0.0:8080"`
	StoreInterval  time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StorePath      string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	RestoreOnStart bool          `env:"RESTORE" envDefault:"true"`
}

func NewServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.StorePath) == 0 && cfg.RestoreOnStart {
		return nil, errors.New("State restoration was requested, but path to store file is not set!")
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

	return sb.String()
}

type Server struct {
	Config     *ServerConfig
	Storage    storage.Storage
	HttpServer *http.Server
}

func NewServer() *Server {
	cfg, err := NewServerConfig()
	if err != nil {
		logging.Log.Fatal(err)
	}

	logging.Log.Info(cfg)

	dataStore := storage.NewDataStore(cfg.StorePath, cfg.StoreInterval)

	recorder := services.NewMetricsRecorder(dataStore)
	router := handlers.Router("./web/views", recorder)
	srv := &http.Server{Addr: cfg.ListenAddress, Handler: router}

	return &Server{
		Config:     cfg,
		Storage:    dataStore,
		HttpServer: srv,
	}
}

func (app *Server) restoreStorage() {
	fileStore, ok := app.Storage.(*storage.FileBackedStorage)
	if !ok {
		logging.Log.Warning("Metrics storage backend doesn't support restoring from disk!")
		return
	}

	if err := fileStore.Restore(); err != nil {
		logging.Log.Fatal(err)
	}
}

func (app *Server) dumpStorage(ctx context.Context) {
	if _, ok := app.Storage.(*storage.FileBackedStorage); !ok {
		logging.Log.Warning("Metrics storage backend doesn't support saving to disk!")
		return
	}

	ticker := time.NewTicker(app.Config.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer tryRecover()

				if err := app.Storage.(*storage.FileBackedStorage).Dump(); err != nil {
					logging.Log.Error(err)
				}
			}()

		case <-ctx.Done():
			logging.Log.Info("Shutdown storage dumping")
			return
		}
	}
}

func (app *Server) Serve(ctx context.Context) {
	if app.Config.RestoreOnStart {
		app.restoreStorage()
	}

	if len(app.Config.StorePath) > 0 && app.Config.StoreInterval > 0 {
		go app.dumpStorage(ctx)
	}

	if err := app.HttpServer.ListenAndServe(); err != http.ErrServerClosed {
		logging.Log.Fatal(err)
	}
}

func (app *Server) Shutdown(signal os.Signal) {
	logging.Log.Info(fmt.Sprintf("Signal '%s' received, shutting down...", signal))

	if err := app.HttpServer.Shutdown(context.Background()); err != nil {
		logging.Log.Error(err)
	}

	dataStore, ok := app.Storage.(*storage.FileBackedStorage)
	if ok {
		if err := dataStore.Dump(); err != nil {
			logging.Log.Error(err)
		}
	}

	logging.Log.Info("Successfully shutdown the service")
}
