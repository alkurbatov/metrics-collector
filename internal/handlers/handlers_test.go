package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sendTestRequest(t *testing.T, method, path string) *http.Response {
	srv := httptest.NewServer(Router(services.RecorderMock{}))
	defer srv.Close()

	req, err := http.NewRequest(method, srv.URL+path, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func TestUpdateMetricHandler(t *testing.T) {
	type result struct {
		code int
	}

	tt := []struct {
		name     string
		path     string
		expected result
	}{
		{
			name: "Push counter",
			path: "/update/counter/PollCount/10",
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Push gauge",
			path: "/update/gauge/Alloc/13.123",
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Push unknown metric kind",
			path: "/update/unknown/Alloc/12.123",
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name: "Push counter with invalid name",
			path: "/update/counter/X)/10",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push gauge with invalid name",
			path: "/update/gauge/X;/10.234",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push counter with invalid value",
			path: "/update/counter/fail/10.0",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push gauge with invalid value",
			path: "/update/gauge/fail/10.234;",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			resp := sendTestRequest(t, http.MethodPost, tc.path)

			assert.Equal(tc.expected.code, resp.StatusCode)

			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Zero(len(respBody))
			}
		})
	}
}

func TestGetMetricHandler(t *testing.T) {
	type result struct {
		code int
	}

	tt := []struct {
		name     string
		path     string
		expected result
	}{
		{
			name: "Get counter",
			path: "/value/counter/PollCount",
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Get gauge",
			path: "/value/gauge/Alloc",
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Get unknown metric kind",
			path: "/value/unknown/Alloc",
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name: "Get unknown counter",
			path: "/value/counter/unknown",
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get unknown gauge",
			path: "/value/gauge/unknown",
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get counter with invalid name",
			path: "/value/counter/X)",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get gauge with invalid name",
			path: "/value/gauge/X;",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			resp := sendTestRequest(t, http.MethodGet, tc.path)

			assert.Equal(tc.expected.code, resp.StatusCode)
			assert.Equal("text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.NotZero(len(respBody))
		})
	}
}

func TestRootHandler(t *testing.T) {
	assert := assert.New(t)

	resp := sendTestRequest(t, http.MethodGet, "/")

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotZero(len(respBody))
}
