package storage_test

import (
	"context"
	"os"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/require"
)

func createStoreWithData(
	t *testing.T,
	storePath string,
	syncMode bool,
) *storage.FileBackedStorage {
	t.Helper()

	ctx := context.Background()
	store := storage.NewFileBackedStorage(storePath, syncMode)

	batch := map[string]storage.Record{
		"PollCount_counter": {Name: "PollCount", Value: metrics.Counter(10)},
		"Alloc_gauge":       {Name: "Alloc", Value: metrics.Gauge(11.356)},
	}
	err := store.PushBatch(ctx, batch)
	require.NoError(t, err)

	err = store.Push(
		ctx,
		"HeapSys_gauge",
		storage.Record{Name: "HeapSys", Value: metrics.Gauge(7831552)},
	)
	require.NoError(t, err)

	err = store.Push(
		ctx,
		"RequestsCount_counter",
		storage.Record{Name: "RequestsCount", Value: metrics.Counter(1)},
	)
	require.NoError(t, err)

	return store
}

func TestSyncDumpRestoreStorage(t *testing.T) {
	storePath := "/tmp/test-sync-dump-restore.json"

	t.Cleanup(func() {
		err := os.Remove(storePath)
		require.NoError(t, err)
	})

	store := createStoreWithData(t, storePath, true)
	storedData := store.Snapshot()

	store = storage.NewFileBackedStorage(storePath, true)
	err := store.Restore()
	require.NoError(t, err)

	restoredData := store.Snapshot()
	require.Equal(t, storedData, restoredData)
}

func TestAsyncDumpRestoreStorage(t *testing.T) {
	storePath := "/tmp/test-async-dump-restore.json"

	t.Cleanup(func() {
		err := os.Remove(storePath)
		require.NoError(t, err)
	})

	store := createStoreWithData(t, storePath, false)
	storedData := store.Snapshot()

	err := store.Close(context.Background())
	require.NoError(t, err)

	store = storage.NewFileBackedStorage(storePath, false)
	err = store.Restore()
	require.NoError(t, err)

	restoredData := store.Snapshot()
	require.Equal(t, storedData, restoredData)
}

func TestRestoreDoesntFailIfNoSourceFile(t *testing.T) {
	store := storage.NewFileBackedStorage("xxx", false)

	err := store.Restore()
	require.NoError(t, err)
}
