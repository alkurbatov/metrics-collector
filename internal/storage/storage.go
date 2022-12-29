package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Storage interface {
	Push(key string, record Record) error
	Get(key string) (Record, bool)
	GetAll() []Record
	String() string
	Close() error
}

// New storage is picked according to the following priority:
// - if DB connection was initialized, use database storage;
// - if filePath is set, use file backed storage;
// - otherwise store data in memory.
func NewDataStore(db *pgx.Conn, filePath string, storeInterval time.Duration) Storage {
	if db != nil {
		return NewDatabaseStorage(db)
	}

	if len(filePath) == 0 {
		return NewMemStorage()
	}

	return NewFileBackedStorage(filePath, storeInterval == 0)
}

type DBConn interface {
	Config() *pgx.ConnConfig
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}
