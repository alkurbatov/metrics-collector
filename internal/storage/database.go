package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var _ Storage = DatabaseStorage{}

func rollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		log.Ctx(ctx).Error().Err(err).Msg("DatabaseStorage - rollback - tx.Rollback")
	}
}

// DatabaseStorage implements database metrics storage.
type DatabaseStorage struct {
	pool DBConnPool
}

// NewDatabaseStorage creates new instance of DatabaseStorage.
func NewDatabaseStorage(pool DBConnPool) DatabaseStorage {
	return DatabaseStorage{pool: pool}
}

// Push records metric data.
func (d DatabaseStorage) Push(ctx context.Context, key string, record Record) error {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("DatabaseStorage - Push - d.pool.Acquire: %w", err)
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		conn.Release()
		return fmt.Errorf("DatabaseStorage - Push - conn.BeginTx: %w", err)
	}

	defer conn.Release()
	defer rollback(ctx, tx)

	if _, err = tx.Exec(
		ctx,
		"INSERT INTO metrics(id, name, kind, value) values ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET value = $4",
		key,
		record.Name,
		record.Value.Kind(),
		record.Value.String(),
	); err != nil {
		return fmt.Errorf("DatabaseStorage - Push - tx.Exec: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("DatabaseStorage - Push - tx.Commit: %w", err)
	}

	return nil
}

// PushBatch records list of metrics data in single request to the database.
func (d DatabaseStorage) PushBatch(ctx context.Context, data map[string]Record) error {
	// NB (alkurbatov): Since batch queries are run in an implicit transaction
	// (unless explicit transaction control statements are executed)
	// we don't need to handle transactions manually.
	// See: https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-EXT-QUERY
	batch := new(pgx.Batch)
	for id, record := range data {
		batch.Queue(
			"INSERT INTO metrics(id, name, kind, value) values ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET value = $4",
			id,
			record.Name,
			record.Value.Kind(),
			record.Value.String(),
		)
	}

	batchResp := d.pool.SendBatch(ctx, batch)
	defer func() {
		if err := batchResp.Close(); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("DatabaseStorage - PushBatch - batchResp.Close")
		}
	}()

	for i := 0; i < len(data); i++ {
		if _, err := batchResp.Exec(); err != nil {
			return fmt.Errorf("DatabaseStorage - PushBatch - batchResp.Exec: %w", err)
		}
	}

	return nil
}

// Get returns stored metrics record.
func (d DatabaseStorage) Get(ctx context.Context, key string) (Record, error) {
	var (
		name  string
		kind  string
		value float64
	)

	err := d.pool.
		QueryRow(ctx, "SELECT name, kind, value FROM metrics WHERE id=$1", key).
		Scan(&name, &kind, &value)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Record{}, fmt.Errorf("DatabaseStorage - Get - d.pool.QueryRow: %w", entity.ErrMetricNotFound)
		}

		return Record{}, fmt.Errorf("DatabaseStorage - Get - d.pool.QueryRow: %w", err)
	}

	switch kind {
	case metrics.KindCounter:
		return Record{Name: name, Value: metrics.Counter(value)}, nil

	case metrics.KindGauge:
		return Record{Name: name, Value: metrics.Gauge(value)}, nil

	default:
		return Record{}, fmt.Errorf("DatabaseStorage - Get - kind: %w", entity.MetricNotImplementedError(kind))
	}
}

// GetAll returns all stored metrics.
func (d DatabaseStorage) GetAll(ctx context.Context) ([]Record, error) {
	rows, err := d.pool.Query(ctx, "SELECT name, kind, value FROM metrics")
	if err != nil {
		return nil, fmt.Errorf("DatabaseStorage - GetAll - d.pool.Query: %w", err)
	}
	defer rows.Close()

	var (
		name  string
		kind  string
		value float64
	)

	rv := make([]Record, 0)
	_, err = pgx.ForEachRow(rows, []any{&name, &kind, &value}, func() error {
		switch kind {
		case metrics.KindCounter:
			rv = append(rv, Record{Name: name, Value: metrics.Counter(value)})
			return nil

		case metrics.KindGauge:
			rv = append(rv, Record{Name: name, Value: metrics.Gauge(value)})
			return nil

		default:
			return entity.MetricNotImplementedError(kind)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("DatabaseStorage - GetAll - pgx.ForEachRow: %w", err)
	}

	return rv, nil
}

// Ping verifies that connection to the database can be established.
func (d DatabaseStorage) Ping(ctx context.Context) error {
	if err := d.pool.Ping(ctx); err != nil {
		return fmt.Errorf("DatabaseSrorage - Ping - d.pool.Ping: %w", err)
	}

	return nil
}

// Close closes all open connection to the database.
func (d DatabaseStorage) Close(_ context.Context) error {
	d.pool.Close()
	return nil
}
