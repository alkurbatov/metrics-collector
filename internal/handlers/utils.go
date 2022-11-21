package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
)

func buildResponse(code int, msg string) string {
	return fmt.Sprintf("%d %s", code, msg)
}

func codeToResponse(code int) string {
	return buildResponse(code, http.StatusText(code))
}

func toCounter(value string) (metrics.Counter, error) {
	rawValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}

	return metrics.Counter(rawValue), nil
}

func toGauge(value string) (metrics.Gauge, error) {
	rawValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return metrics.Gauge(rawValue), nil
}
