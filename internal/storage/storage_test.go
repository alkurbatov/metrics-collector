package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDataStore(t *testing.T) {
	tt := []struct {
		name     string
		path     string
		interval time.Duration
		expected Storage
	}{
		{
			name:     "Path and interval are set",
			path:     "some/path",
			interval: 10 * time.Second,
			expected: &FileBackedStorage{},
		},
		{
			name:     "Interval is zero",
			path:     "some/path",
			interval: 0,
			expected: &FileBackedStorage{},
		},
		{
			name:     "Path not set",
			path:     "",
			interval: 10,
			expected: &MemStorage{},
		},
		{
			name:     "Path not set and interval is zero",
			path:     "",
			interval: 0,
			expected: &MemStorage{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			store := NewDataStore(tc.path, tc.interval)
			assert.IsType(t, tc.expected, store)
		})
	}
}
