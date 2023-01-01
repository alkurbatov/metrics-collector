package services

import (
	"context"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

type HealthCheckImpl struct {
	storage storage.Storage
}

func NewHealthCheck(dataStore storage.Storage) HealthCheckImpl {
	return HealthCheckImpl{storage: dataStore}
}

func (h HealthCheckImpl) CheckStorage(ctx context.Context) error {
	database, ok := h.storage.(storage.DatabaseStorage)
	if !ok {
		return entity.ErrHealthCheckNotSupported
	}

	return database.Ping(ctx)
}
