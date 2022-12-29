package schema_test

import (
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestValidateMetricsName(t *testing.T) {
	tt := []struct {
		name   string
		metric string
		valid  bool
	}{
		{
			name:   "Basic name",
			metric: "Alloc",
			valid:  true,
		},
		{
			name:   "Long name with several capital letters",
			metric: "NumForcedGC",
			valid:  true,
		},
		{
			name:   "Name with lowercase letters",
			metric: "count",
			valid:  true,
		},
		{
			name:   "Name with numbers at the beginning",
			metric: "1Num",
			valid:  true,
		},
		{
			name:   "Name with numbers in middle",
			metric: "Num5Num",
			valid:  true,
		},
		{
			name:   "Name with numbers at the end",
			metric: "Num1",
			valid:  true,
		},
		{
			name:   "Name only with numbers",
			metric: "123",
			valid:  true,
		},
		{
			name:   "Name with a dot",
			metric: "Some.Name",
			valid:  false,
		},
		{
			name:   "Name with a -",
			metric: "Some-Name",
			valid:  false,
		},
		{
			name:   "Name with a _",
			metric: "Some_Name",
			valid:  false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := schema.ValidateMetricName(tc.metric)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
