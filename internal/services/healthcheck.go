package services

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

func storageCheckError(reason error) error {
	return fmt.Errorf("storage check failed: %w", reason)
}

type HealthCheckImpl struct {
	storage storage.Storage
}

func NewHealthCheck(dataStore storage.Storage) HealthCheckImpl {
	return HealthCheckImpl{storage: dataStore}
}

func (h HealthCheckImpl) CheckStorage(ctx context.Context) error {
	database, ok := h.storage.(storage.DatabaseStorage)
	if !ok {
		return storageCheckError(entity.ErrHealthCheckNotSupported)
	}

	if err := database.Ping(ctx); err != nil {
		return storageCheckError(err)
	}

	return nil
}
