package services_test

import (
	"context"
	"errors"
	"testing"

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
			err:  errors.New("offline"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := storage.NewDBConnMock()
			m.On("Ping", mock.Anything).Return(tc.err)

			store := storage.NewDatabaseStorage(m)
			probe := services.NewHealthCheck(store)

			err := probe.CheckStorage(context.Background())
			assert.ErrorIs(t, tc.err, err)
		})
	}
}

func TestCheckStorageOnUnsupportedStorage(t *testing.T) {
	store := storage.NewMemStorage()
	probe := services.NewHealthCheck(store)

	err := probe.CheckStorage(context.Background())
	assert.ErrorIs(t, err, services.ErrHealthCheckNotSupported)
}
