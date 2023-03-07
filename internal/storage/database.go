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

func pushError(reason error) error {
	return fmt.Errorf("failed to push record to DB: %w", reason)
}

func pushListError(reason error) error {
	return fmt.Errorf("failed to push records list to DB: %w", reason)
}

func getError(reason error) error {
	return fmt.Errorf("failed to get record from DB: %w", reason)
}

func getListError(reason error) error {
	return fmt.Errorf("failed to get records list from DB: %w", reason)
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
		return pushError(err)
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		conn.Release()
		return pushError(err)
	}

	defer conn.Release()
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Ctx(ctx).Error().Err(pushError(err)).Msg("")
		}
	}()

	if _, err = tx.Exec(
		ctx,
		"INSERT INTO metrics(id, name, kind, value) values ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET value = $4",
		key,
		record.Name,
		record.Value.Kind(),
		record.Value.String(),
	); err != nil {
		return pushError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return pushError(err)
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
			log.Ctx(ctx).Error().Err(pushListError(err)).Msg("")
		}
	}()

	for i := 0; i < len(data); i++ {
		if _, err := batchResp.Exec(); err != nil {
			return pushListError(err)
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
			return Record{}, getError(entity.ErrMetricNotFound)
		}

		return Record{}, getError(err)
	}

	switch kind {
	case metrics.KindCounter:
		return Record{Name: name, Value: metrics.Counter(value)}, nil

	case metrics.KindGauge:
		return Record{Name: name, Value: metrics.Gauge(value)}, nil

	default:
		return Record{}, getError(entity.MetricNotImplementedError(kind))
	}
}

// GetAll returns all stored metrics.
func (d DatabaseStorage) GetAll(ctx context.Context) ([]Record, error) {
	rows, err := d.pool.Query(ctx, "SELECT name, kind, value FROM metrics")
	if err != nil {
		return nil, getListError(err)
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
			return entity.MetricNotImplementedError(kind) //nolint: wrapcheck
		}
	})

	if err != nil {
		return nil, getListError(err)
	}

	return rv, nil
}

// Ping verifies that connection to the database can be established.
func (d DatabaseStorage) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx) //nolint: wrapcheck
}

// Close closes all open connection to the database.
func (d DatabaseStorage) Close() error {
	d.pool.Close()
	return nil
}
