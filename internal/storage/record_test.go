package storage_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestRecordConvertedToJsonAndBack(t *testing.T) {
	tt := []struct {
		name      string
		srcRecord storage.Record
	}{
		{
			name: "Counter",
			srcRecord: storage.Record{
				Name:  "PollCount",
				Value: metrics.Counter(10),
			},
		},
		{
			name: "Gauge",
			srcRecord: storage.Record{
				Name:  "Alloc",
				Value: metrics.Gauge(111.456789),
			},
		},
		{
			name: "Gauge with zero mantissa",
			srcRecord: storage.Record{
				Name:  "Alloc",
				Value: metrics.Gauge(111.0),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			json, err := tc.srcRecord.MarshalJSON()
			require.NoError(err)

			dstRecord := new(storage.Record)
			err = dstRecord.UnmarshalJSON(json)
			require.NoError(err)

			require.Equal(&tc.srcRecord, dstRecord)
		})
	}
}
