package schema

import (
	"regexp"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/services"
)

var metricName = regexp.MustCompile(`^[A-Za-z\d]+$`)

func ValidateMetricName(name, kind string) error {
	if len(services.CalculateID(name, kind)) > 255 {
		return entity.ErrMetricLongName
	}

	if !metricName.MatchString(name) {
		return entity.ErrMetricInvalidName
	}

	return nil
}

func ValidateMetricKind(kind string) error {
	switch kind {
	case entity.Counter, entity.Gauge:
		return nil

	default:
		return entity.MetricNotImplementedError(kind)
	}
}
