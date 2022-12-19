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

	flag "github.com/spf13/pflag"
)

type ServerConfig struct {
	ListenAddress  NetAddress    `env:"ADDRESS"`
	StoreInterval  time.Duration `env:"STORE_INTERVAL"`
	StorePath      string        `env:"STORE_FILE"`
	RestoreOnStart bool          `env:"RESTORE"`
}

func NewServerConfig() (*ServerConfig, error) {
	listenAddress := NetAddress("0.0.0.0:8080")
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
		true, "whether to restore state on startup or not",
	)
	flag.Parse()

	cfg := &ServerConfig{
		ListenAddress:  listenAddress,
		StorePath:      *storePath,
		StoreInterval:  *storeInterval,
		RestoreOnStart: *restoreOnStart,
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.StorePath) == 0 && cfg.RestoreOnStart {
		return nil, errors.New("state restoration was requested, but path to store file is not set")
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
	HTTPServer *http.Server
}

func NewServer() *Server {
	cfg, err := NewServerConfig()
	if err != nil {
		logging.Log.Fatal(err)
	}

	logging.Log.Info(cfg)

	dataStore := storage.NewDataStore(cfg.StorePath, cfg.StoreInterval)
	logging.Log.Info("Attached " + dataStore.String())

	recorder := services.NewMetricsRecorder(dataStore)
	router := handlers.Router("./web/views", recorder)
	srv := &http.Server{Addr: cfg.ListenAddress.String(), Handler: router}

	return &Server{
		Config:     cfg,
		Storage:    dataStore,
		HTTPServer: srv,
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

	if err := app.HTTPServer.ListenAndServe(); err != http.ErrServerClosed {
		logging.Log.Fatal(err)
	}
}

func (app *Server) Shutdown(signal os.Signal) {
	logging.Log.Info(fmt.Sprintf("Signal '%s' received, shutting down...", signal))

	if err := app.HTTPServer.Shutdown(context.Background()); err != nil {
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
