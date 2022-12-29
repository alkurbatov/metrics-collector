package services

import (
	"context"
	"errors"

	"github.com/alkurbatov/metrics-collector/internal/storage"
)

var ErrHealthCheckNotSupported = errors.New("storage doesn't support healthcheck")

type HealthCheckImpl struct {
	storage storage.Storage
}

func NewHealthCheck(dataStore storage.Storage) HealthCheckImpl {
	return HealthCheckImpl{storage: dataStore}
}

func (h HealthCheckImpl) CheckStorage(ctx context.Context) error {
	database, ok := h.storage.(*storage.DatabaseStorage)
	if !ok {
		return ErrHealthCheckNotSupported
	}

	return database.Ping(ctx)
}
