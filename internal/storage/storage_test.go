package storage_test

import (
	"testing"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestNewDataStore(t *testing.T) {
	tt := []struct {
		name     string
		db       *pgxpool.Pool
		path     string
		interval time.Duration
		expected storage.Storage
	}{
		{
			name:     "Should create database storage, if DB connection set",
			path:     "some/path",
			db:       &pgxpool.Pool{},
			interval: 10 * time.Second,
			expected: storage.DatabaseStorage{},
		},
		{
			name:     "Should create file backed storage, if path and interval are set",
			path:     "some/path",
			interval: 10 * time.Second,
			expected: &storage.FileBackedStorage{},
		},
		{
			name:     "Should create file backed storage, if path set without interval",
			path:     "some/path",
			interval: 0,
			expected: &storage.FileBackedStorage{},
		},
		{
			name:     "Should create memory storage, if path not set",
			path:     "",
			interval: 10,
			expected: &storage.MemStorage{},
		},
		{
			name:     "Should create memory storage, if path not set and interval is zero",
			path:     "",
			interval: 0,
			expected: &storage.MemStorage{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			store := storage.NewDataStore(tc.db, tc.path, tc.interval)
			assert.IsType(t, tc.expected, store)
		})
	}
}
