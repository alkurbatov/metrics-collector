// Package services contains implementationm of business logic for different scenarios.
package services

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type Recorder interface {
	Push(ctx context.Context, record storage.Record) (storage.Record, error)
	PushList(ctx context.Context, records []storage.Record) error
	Get(ctx context.Context, kind, name string) (storage.Record, error)
	List(ctx context.Context) ([]storage.Record, error)
}

type HealthCheck interface {
	CheckStorage(ctx context.Context) error
}
