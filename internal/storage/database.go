package storage

import (
	"context"
	"errors"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/logging"
	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/jackc/pgx/v5"
)

func rollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		logging.Log.Error(err)
	}
}

type DatabaseStorage struct {
	pool DBConnPool
}

func NewDatabaseStorage(pool DBConnPool) DatabaseStorage {
	return DatabaseStorage{pool: pool}
}

func (d DatabaseStorage) Push(ctx context.Context, key string, record Record) error {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return err
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
		return err
	}

	return tx.Commit(ctx)
}

func (d DatabaseStorage) Get(ctx context.Context, key string) (*Record, error) {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Release()

	var (
		name  string
		kind  string
		value float64
	)

	err = conn.QueryRow(ctx, "SELECT name, kind, value FROM metrics WHERE id=$1", key).Scan(&name, &kind, &value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrMetricNotFound
		}

		return nil, err
	}

	switch kind {
	case entity.Counter:
		return &Record{Name: name, Value: metrics.Counter(value)}, nil

	case entity.Gauge:
		return &Record{Name: name, Value: metrics.Gauge(value)}, nil

	default:
		return nil, &entity.MetricNotImplementedError{Kind: kind}
	}
}

func (d DatabaseStorage) GetAll(ctx context.Context) ([]Record, error) {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT name, kind, value FROM metrics")
	if err != nil {
		return nil, err
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
		case entity.Counter:
			rv = append(rv, Record{Name: name, Value: metrics.Counter(value)})
			return nil

		case entity.Gauge:
			rv = append(rv, Record{Name: name, Value: metrics.Gauge(value)})
			return nil

		default:
			return &entity.MetricNotImplementedError{Kind: kind}
		}
	})

	return rv, err
}

func (d DatabaseStorage) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}

func (d DatabaseStorage) Close() error {
	d.pool.Close()
	return nil
}
