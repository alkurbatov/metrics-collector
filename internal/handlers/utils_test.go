package handlers

import (
	"net/http"
	"testing"

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
