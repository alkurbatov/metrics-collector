package handlers

import (
	"errors"
	"regexp"
)

var metricName *regexp.Regexp = regexp.MustCompile(`^[A-Za-z\d]+$`)

func validateMetricName(name string) error {
	if metricName.MatchString(name) {
		return nil
	}

	return errors.New("metrics name contains invalid characters")
}
