package storage_test

import (
	"context"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPing(t *testing.T) {
	tt := []struct {
		name   string
		result error
	}{
		{
			name: "Should return no error if DB is online",
		},
		{
			name:   "Should return error if DB is ofline",
			result: entity.ErrUnexpected,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := storage.NewDBConnPoolMock()
			m.On("Ping", mock.Anything).Return(tc.result)

			s := storage.NewDatabaseStorage(m)
			err := s.Ping(context.Background())
			assert.ErrorIs(t, err, tc.result)
		})
	}
}

func TestCloseNeverFails(t *testing.T) {
	m := storage.NewDBConnPoolMock()
	m.On("Close").Return()

	s := storage.NewDatabaseStorage(m)
	assert.NoError(t, s.Close(context.Background()))
}
