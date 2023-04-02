package services

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/storage"
)

var _ HealthCheck = (*HealthCheckImpl)(nil)

func storageCheckError(reason error) error {
	return fmt.Errorf("storage check failed: %w", reason)
}

// HealthCheckImpl implements general service healthcheck.
// Currently only database connection is verified.
type HealthCheckImpl struct {
	storage storage.Storage
}

func NewHealthCheck(dataStore storage.Storage) HealthCheckImpl {
	return HealthCheckImpl{storage: dataStore}
}

// CheckStorage verifies connection to the database.
// Fails if database storage is not configured.
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
