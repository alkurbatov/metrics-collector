package storage_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
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
			name: "Should convert counter",
			srcRecord: storage.Record{
				Name:  "PollCount",
				Value: metrics.Counter(10),
			},
		},
		{
			name: "Should convert gauge",
			srcRecord: storage.Record{
				Name:  "Alloc",
				Value: metrics.Gauge(111.456789),
			},
		},
		{
			name: "Should convert gauge with zero mantissa",
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

func TestUmrashalJSONOnCorruptedData(t *testing.T) {
	tt := []struct {
		name     string
		data     string
		expected error
	}{
		{
			name:     "Should fail on malformed JSON",
			data:     `{"name": "xxx",`,
			expected: &json.SyntaxError{},
		},
		{
			name:     "Should fail on bad counter",
			data:     `{"name": "xxx", "kind": "counter", "value": "12.345"}`,
			expected: strconv.ErrSyntax,
		},
		{
			name:     "Should fail on bad gauge",
			data:     `{"name": "xxx", "kind": "gauge", "value": "12.)"}`,
			expected: strconv.ErrSyntax,
		},
		{
			name:     "Should fail on unknown counter",
			data:     `{"name": "xxx", "kind": "unknown", "value": "12"}`,
			expected: &entity.MetricNotImplementedError{Kind: "unknown"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := new(storage.Record)

			err := r.UnmarshalJSON([]byte(tc.data))
			require.ErrorAs(t, err, &tc.expected)
		})
	}
}
