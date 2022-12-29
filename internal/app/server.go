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
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5"

	flag "github.com/spf13/pflag"
)

type ServerConfig struct {
	ListenAddress  NetAddress           `env:"ADDRESS"`
	StoreInterval  time.Duration        `env:"STORE_INTERVAL"`
	StorePath      string               `env:"STORE_FILE"`
	RestoreOnStart bool                 `env:"RESTORE"`
	Secret         security.Secret      `env:"KEY"`
	DatabaseDSN    security.DatabaseURL `env:"DATABASE_DSN"`
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
		true,
		"whether to restore state on startup or not",
	)
	secret := flag.StringP(
		"key",
		"k",
		"",
		"secret key for signature generation",
	)
	databaseDSN := flag.StringP(
		"db-dsn",
		"d",
		"",
		"full database connection URL",
	)

	flag.Parse()

	cfg := &ServerConfig{
		ListenAddress:  listenAddress,
		StorePath:      *storePath,
		StoreInterval:  *storeInterval,
		RestoreOnStart: *restoreOnStart,
		Secret:         security.Secret(*secret),
		DatabaseDSN:    security.DatabaseURL(*databaseDSN),
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

	if len(c.Secret) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tSecret key: %s\n", c.Secret))
	}

	if len(c.DatabaseDSN) > 0 {
		sb.WriteString(fmt.Sprintf("\t\tDatabase DSN: %s\n", c.DatabaseDSN))
	}

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

	var db *pgx.Conn
	if len(cfg.DatabaseDSN) > 0 {
		db, err = pgx.Connect(context.Background(), string(cfg.DatabaseDSN))
		if err != nil {
			logging.Log.Fatal(err)
		}
	}

	dataStore := storage.NewDataStore(db, cfg.StorePath, cfg.StoreInterval)
	logging.Log.Info("Attached " + dataStore.String())

	recorder := services.NewMetricsRecorder(dataStore)
	healthcheck := services.NewHealthCheck(dataStore)

	var signer *security.Signer
	if len(cfg.Secret) > 0 {
		signer = security.NewSigner(cfg.Secret)
	}

	router := handlers.Router("./web/views", recorder, healthcheck, signer)
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
	if len(app.Config.DatabaseDSN) == 0 && len(app.Config.StorePath) > 0 {
		if app.Config.RestoreOnStart {
			app.restoreStorage()
		}

		if app.Config.StoreInterval > 0 {
			go app.dumpStorage(ctx)
		}
	}

	if err := app.HTTPServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logging.Log.Fatal(err)
	}
}

func (app *Server) Shutdown(signal os.Signal) {
	logging.Log.Info(fmt.Sprintf("Signal '%s' received, shutting down...", signal))

	if err := app.HTTPServer.Shutdown(context.Background()); err != nil {
		logging.Log.Error(err)
	}

	if err := app.Storage.Close(); err != nil {
		logging.Log.Error(err)
	}

	logging.Log.Info("Successfully shutdown the service")
}
