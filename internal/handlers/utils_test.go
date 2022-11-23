package handlers

import (
	"net/http"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestBuildResponse(t *testing.T) {
	type args struct {
		code int
		msg  string
	}

	tt := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name: "Format HTTP code with custom message",
			args: args{
				code: http.StatusNotFound,
				msg:  "Not found",
			},
			expected: "404 Not found",
		},
		{
			name: "Format generic code with a message",
			args: args{
				code: 15,
				msg:  "Unknown error",
			},
			expected: "15 Unknown error",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp := buildResponse(tc.args.code, tc.args.msg)
			assert.Equal(t, tc.expected, resp)
		})
	}
}

func TestCodeToResponse(t *testing.T) {
	tt := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "Format HTTP",
			code:     http.StatusNotFound,
			expected: "404 Not Found",
		},
		{
			name:     "Format unexisting HTTP code",
			code:     1000,
			expected: "1000 ",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp := codeToResponse(tc.code)
			assert.Equal(t, tc.expected, resp)
		})
	}
}

func TestToCounter(t *testing.T) {
	tt := []struct {
		name     string
		value    string
		valid    bool
		expected metrics.Counter
	}{
		{
			name:     "Positive integer",
			value:    "15",
			valid:    true,
			expected: 15,
		},
		{
			name:     "Zero integer",
			value:    "0",
			valid:    true,
			expected: 0,
		},
		{
			name:     "Negative integer",
			value:    "-15",
			valid:    true,
			expected: -15,
		},
		{
			name:  "Positive float",
			value: "28968.000000",
			valid: false,
		},
		{
			name:  "Zero float",
			value: "0.000000",
			valid: false,
		},
		{
			name:  "Negative float",
			value: "-28968.000000",
			valid: false,
		},
	}

	assert := assert.New(t)
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			metric, err := toCounter(tc.value)
			if tc.valid {
				assert.NoError(err)
				assert.Equal(tc.expected, metric)
			} else {
				assert.NotNil(err)
			}
		})
	}
}

func TestToGauge(t *testing.T) {
	tt := []struct {
		name     string
		value    string
		valid    bool
		expected metrics.Gauge
	}{
		{
			name:     "Positive integer",
			value:    "15",
			valid:    true,
			expected: 15.0,
		},
		{
			name:     "Zero integer",
			value:    "0",
			valid:    true,
			expected: 0.0,
		},
		{
			name:     "Negative integer",
			value:    "-15",
			valid:    true,
			expected: -15.0,
		},
		{
			name:     "Positive float",
			value:    "28968.134000",
			valid:    true,
			expected: 28968.134,
		},
		{
			name:     "Zero float",
			value:    "0.000000",
			valid:    true,
			expected: 0.0,
		},
		{
			name:     "Negative float",
			value:    "-28968.134000",
			valid:    true,
			expected: -28968.134,
		},
		{
			name:     "Small positive float",
			value:    "0.604660",
			valid:    true,
			expected: 0.604660,
		},
		{
			name:  "Meaningless value",
			value: "...",
			valid: false,
		},
		{
			name:  "Malformed value",
			value: "0.12(",
			valid: false,
		},
	}

	assert := assert.New(t)
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			metric, err := toGauge(tc.value)
			if tc.valid {
				assert.NoError(err)
				assert.Equal(tc.expected, metric)
			} else {
				assert.NotNil(err)
			}
		})
	}
}
