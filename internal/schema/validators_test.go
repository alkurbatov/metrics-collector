package schema_test

import (
	"strings"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestValidateMetricsName(t *testing.T) {
	tt := []struct {
		name   string
		metric string
		kind   string
		err    error
	}{
		{
			name:   "Should accept basic name",
			metric: "Alloc",
			kind:   entity.Counter,
		},
		{
			name:   "Should accept long name with several capital letters",
			metric: "NumForcedGC",
			kind:   entity.Gauge,
		},
		{
			name:   "Should accept name with lowercase letters",
			metric: "count",
			kind:   entity.Counter,
		},
		{
			name:   "Should accept name with numbers at the beginning",
			metric: "1Num",
			kind:   entity.Gauge,
		},
		{
			name:   "Should accept name with numbers in middle",
			metric: "Num5Num",
			kind:   entity.Gauge,
		},
		{
			name:   "Should accept name with numbers at the end",
			metric: "Num1",
			kind:   entity.Gauge,
		},
		{
			name:   "Should accept name only with numbers",
			metric: "123",
			kind:   entity.Gauge,
		},
		{
			name:   "Should reject name with a dot",
			metric: "Some.Name",
			kind:   entity.Gauge,
			err:    entity.ErrMetricInvalidName,
		},
		{
			name:   "Should reject name with a -",
			metric: "Some-Name",
			kind:   entity.Gauge,
			err:    entity.ErrMetricInvalidName,
		},
		{
			name:   "Should reject name with a _",
			metric: "Some_Name",
			kind:   entity.Gauge,
			err:    entity.ErrMetricInvalidName,
		},
		{
			name: "Should reject empty names",
			kind: entity.Gauge,
			err:  entity.ErrMetricInvalidName,
		},
		{
			name:   "Should accept long counter names under limit",
			metric: strings.Repeat("a", 254-len(entity.Counter)),
			kind:   entity.Counter,
		},
		{
			name:   "Should accept long gauge names under limit",
			metric: strings.Repeat("a", 254-len(entity.Gauge)),
			kind:   entity.Gauge,
		},
		{
			name:   "Should reject too long names",
			metric: strings.Repeat("a", 250),
			kind:   entity.Gauge,
			err:    entity.ErrMetricLongName,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.ValidateMetricName(tc.metric, tc.kind)
			assert.ErrorIs(t, tc.err, err)
		})
	}
}

func TestValidateMetricKind(t *testing.T) {
	tt := []struct {
		name string
		kind string
		err  error
	}{
		{
			name: "Should accept counter",
			kind: entity.Counter,
		},
		{
			name: "Should accept gauge",
			kind: entity.Gauge,
		},
		{
			name: "Should reject unknown kind",
			kind: "xxx",
			err:  entity.MetricNotImplementedError("xxx"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.ValidateMetricKind(tc.kind)

			if tc.err == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorAs(t, err, &tc.err)
		})
	}
}
