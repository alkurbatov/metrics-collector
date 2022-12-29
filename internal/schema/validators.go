package schema

import (
	"errors"
	"regexp"
)

var metricName = regexp.MustCompile(`^[A-Za-z\d]+$`)

func ValidateMetricName(name string) error {
	if metricName.MatchString(name) {
		return nil
	}

	return errors.New("metrics name contains invalid characters")
}
