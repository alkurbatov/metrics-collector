package services_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckStorage(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{
			name: "Should return nil, if storage online",
		},
		{
			name: "Should return error, if storage offline",
			err:  entity.ErrUnexpected,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := storage.NewDBConnPoolMock()
			m.On("Ping", mock.Anything).Return(tc.err)

			store := storage.NewDatabaseStorage(m)
			probe := services.NewHealthCheck(store)

			err := probe.CheckStorage(context.Background())
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestCheckStorageOnUnsupportedStorage(t *testing.T) {
	store := storage.NewMemStorage()
	probe := services.NewHealthCheck(store)

	err := probe.CheckStorage(context.Background())
	assert.ErrorIs(t, err, entity.ErrHealthCheckNotSupported)
}
