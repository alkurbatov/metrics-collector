package storage

import (
	"context"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Push(ctx context.Context, key string, record Record) error
	PushList(ctx context.Context, keys []string, records []Record) error
	Get(ctx context.Context, key string) (Record, error)
	GetAll(ctx context.Context) ([]Record, error)
	Close() error
}

// New storage is picked according to the following priority:
// - if DB connection was initialized, use database storage;
// - if filePath is set, use file backed storage;
// - otherwise store data in memory.
func NewDataStore(pool *pgxpool.Pool, filePath string, storeInterval time.Duration) Storage {
	if pool != nil {
		logging.Log.Info("Attached database storage")
		return NewDatabaseStorage(pool)
	}

	if len(filePath) == 0 {
		logging.Log.Info("Attached in-memory storage")
		return NewMemStorage()
	}

	logging.Log.Info("Attached file-backed storage")

	return NewFileBackedStorage(filePath, storeInterval == 0)
}

type DBConnPool interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Ping(ctx context.Context) error
	Close()
}
