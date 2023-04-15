// Package validators provides common validators for API requests.
package validators

import (
	"regexp"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
)

var metricName = regexp.MustCompile(`^[A-Za-z\d]+$`)

// ValidateMetricName verifies that provided metric name is acceptable.
func ValidateMetricName(name, kind string) error {
	if len(services.CalculateID(name, kind)) > 255 {
		return entity.ErrMetricLongName
	}

	if !metricName.MatchString(name) {
		return entity.ErrMetricInvalidName
	}

	return nil
}

// ValidateMetricKind verifies that provided metric kind is known.
func ValidateMetricKind(kind string) error {
	switch kind {
	case metrics.KindCounter, metrics.KindGauge:
		return nil

	default:
		return entity.MetricNotImplementedError(kind) //nolint: wrapcheck
	}
}
