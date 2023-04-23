package validators_test

import (
	"strings"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/validators"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
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
			kind:   metrics.KindCounter,
		},
		{
			name:   "Should accept long name with several capital letters",
			metric: "NumForcedGC",
			kind:   metrics.KindGauge,
		},
		{
			name:   "Should accept name with lowercase letters",
			metric: "count",
			kind:   metrics.KindCounter,
		},
		{
			name:   "Should accept name with numbers at the beginning",
			metric: "1Num",
			kind:   metrics.KindGauge,
		},
		{
			name:   "Should accept name with numbers in middle",
			metric: "Num5Num",
			kind:   metrics.KindGauge,
		},
		{
			name:   "Should accept name with numbers at the end",
			metric: "Num1",
			kind:   metrics.KindGauge,
		},
		{
			name:   "Should accept name only with numbers",
			metric: "123",
			kind:   metrics.KindGauge,
		},
		{
			name:   "Should reject name with a dot",
			metric: "Some.Name",
			kind:   metrics.KindGauge,
			err:    entity.ErrMetricInvalidName,
		},
		{
			name:   "Should reject name with a -",
			metric: "Some-Name",
			kind:   metrics.KindGauge,
			err:    entity.ErrMetricInvalidName,
		},
		{
			name:   "Should reject name with a _",
			metric: "Some_Name",
			kind:   metrics.KindGauge,
			err:    entity.ErrMetricInvalidName,
		},
		{
			name: "Should reject empty names",
			kind: metrics.KindGauge,
			err:  entity.ErrMetricInvalidName,
		},
		{
			name:   "Should accept long counter names under limit",
			metric: strings.Repeat("a", 254-len(metrics.KindCounter)),
			kind:   metrics.KindCounter,
		},
		{
			name:   "Should accept long gauge names under limit",
			metric: strings.Repeat("a", 254-len(metrics.KindGauge)),
			kind:   metrics.KindGauge,
		},
		{
			name:   "Should reject too long names",
			metric: strings.Repeat("a", 250),
			kind:   metrics.KindGauge,
			err:    entity.ErrMetricLongName,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateMetricName(tc.metric, tc.kind)
			assert.ErrorIs(t, err, tc.err)
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
			kind: metrics.KindCounter,
		},
		{
			name: "Should accept gauge",
			kind: metrics.KindGauge,
		},
		{
			name: "Should reject unknown kind",
			kind: "xxx",
			err:  entity.MetricNotImplementedError("xxx"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateMetricKind(tc.kind)

			if tc.err == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorAs(t, err, &tc.err)
		})
	}
}

func TestValidateTransport(t *testing.T) {
	tt := []struct {
		name      string
		transport string
		err       error
	}{
		{
			name:      "HTTP transport is supported",
			transport: entity.TransportHTTP,
			err:       nil,
		},
		{
			name:      "gRPC transport is supported",
			transport: entity.TransportGRPC,
			err:       nil,
		},
		{
			name:      "Unknown transport is not supported",
			transport: "unknown",
			err:       entity.ErrTransportNotSupported,
		},
		{
			name:      "Empty transport is not supported",
			transport: "",
			err:       entity.ErrTransportNotSupported,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := validators.ValidateTransport(tc.transport)

			if tc.err == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorAs(t, err, &tc.err)
		})
	}
}
