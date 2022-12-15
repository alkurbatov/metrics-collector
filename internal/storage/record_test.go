package storage

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/stretchr/testify/require"
)

func TestRecordConvertedToJsonAndBack(t *testing.T) {
	tt := []struct {
		name      string
		srcRecord Record
	}{
		{
			name: "Counter",
			srcRecord: Record{
				Name:  "PollCount",
				Value: metrics.Counter(10),
			},
		},
		{
			name: "Gauge",
			srcRecord: Record{
				Name:  "Alloc",
				Value: metrics.Gauge(111.456789),
			},
		},
		{
			name: "Gauge with zero mantissa",
			srcRecord: Record{
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

			dstRecord := new(Record)
			err = dstRecord.UnmarshalJSON(json)
			require.NoError(err)

			require.Equal(&tc.srcRecord, dstRecord)
		})
	}
}
