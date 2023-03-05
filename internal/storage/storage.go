package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Storage interface {
	Push(ctx context.Context, key string, record Record) error
	PushBatch(ctx context.Context, data map[string]Record) error
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
		log.Info().Msg("Attached database storage")
		return NewDatabaseStorage(pool)
	}

	if len(filePath) == 0 {
		log.Info().Msg("Attached in-memory storage")
		return NewMemStorage()
	}

	log.Info().Msg("Attached file-backed storage")

	return NewFileBackedStorage(filePath, storeInterval == 0)
}

type DBConnPool interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Ping(ctx context.Context) error
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Close()
}
