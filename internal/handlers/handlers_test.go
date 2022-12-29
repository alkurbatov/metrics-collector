package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/schema"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sendTestRequest(t *testing.T, method, path string, payload []byte, key security.Secret) *http.Response {
	t.Helper()

	var signer *security.Signer
	if len(key) > 0 {
		signer = security.NewSigner(key)
	}

	srv := httptest.NewServer(Router("../../web/views", services.RecorderMock{}, signer))
	defer srv.Close()

	body := bytes.NewReader(payload)

	req, err := http.NewRequest(method, srv.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func TestUpdateMetric(t *testing.T) {
	type result struct {
		code int
		body string
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
				body: "10",
			},
		},
		{
			name: "Push gauge",
			path: "/update/gauge/Alloc/13.123",
			expected: result{
				code: http.StatusOK,
				body: "13.123",
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
			path: "/update/counter/PollCount/10\\.0",

			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push gauge with invalid value",
			path: "/update/gauge/Alloc/15.234;",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push counter when recorder fails",
			path: "/update/counter/fail/10",
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "Push gauge when recorder fails",
			path: "/update/gauge/fail/10.234",
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			resp := sendTestRequest(t, http.MethodPost, tc.path, nil, "")

			assert.Equal(tc.expected.code, resp.StatusCode)

			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(tc.expected.body, string(respBody))
			}
		})
	}
}

func TestUpdateJSONMetric(t *testing.T) {
	type result struct {
		code int
	}

	tt := []struct {
		name      string
		req       schema.MetricReq
		clientKey security.Secret
		serverKey security.Secret
		expected  result
	}{
		{
			name: "Push counter",
			req:  schema.NewUpdateCounterReq("PollCount", 10),
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name: "Push gauge",
			req:  schema.NewUpdateGaugeReq("Alloc", 13.123),
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:      "Should push counter with signature",
			req:       schema.NewUpdateCounterReq("PollCount", 10),
			clientKey: "abc",
			serverKey: "abc",
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:      "Should push gauge with signature",
			req:       schema.NewUpdateGaugeReq("Alloc", 13.123),
			clientKey: "abc",
			serverKey: "abc",
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:      "Should fail if counter signature doesn't match",
			req:       schema.NewUpdateCounterReq("PollCount", 10),
			clientKey: "abc",
			serverKey: "xxx",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:      "Should fail if gauge signature doesn't match",
			req:       schema.NewUpdateGaugeReq("Alloc", 13.123),
			clientKey: "abc",
			serverKey: "xxx",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push unknown metric kind",
			req: schema.MetricReq{
				ID:    "X",
				MType: "unknown",
			},
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name: "Push counter with invalid name",
			req:  schema.NewUpdateCounterReq("X)", 10),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push gauge with invalid name",
			req:  schema.NewUpdateGaugeReq("X;", 13.123),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Push counter  when recorder fails",
			req:  schema.NewUpdateCounterReq("fail", 13),
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
		{
			name: "Push gauge when recorder fails",
			req:  schema.NewUpdateGaugeReq("fail", 13.123),
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if len(tc.clientKey) > 0 {
				signer := security.NewSigner(tc.clientKey)
				err := signer.SignRequest(&tc.req)
				require.NoError(err)
			}

			payload, err := json.Marshal(tc.req)
			require.NoError(err)

			resp := sendTestRequest(t, http.MethodPost, "/update", payload, tc.serverKey)

			assert.Equal(tc.expected.code, resp.StatusCode)

			if tc.expected.code == http.StatusOK {
				assert.Equal("application/json", resp.Header.Get("Content-Type"))

				respBody, err := io.ReadAll(resp.Body)
				require.NoError(err)
				defer resp.Body.Close()

				var resp schema.MetricReq
				err = json.Unmarshal(respBody, &resp)
				require.NoError(err)

				assert.Equal(tc.req, resp)
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	type result struct {
		code int
		body string
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
				body: "10",
			},
		},
		{
			name: "Get gauge",
			path: "/value/gauge/Alloc",
			expected: result{
				code: http.StatusOK,
				body: "11.345",
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

			resp := sendTestRequest(t, http.MethodGet, tc.path, nil, "")

			assert.Equal(tc.expected.code, resp.StatusCode)
			assert.Equal("text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

			if tc.expected.code == http.StatusOK {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(tc.expected.body, string(respBody))
			}
		})
	}
}

func TestGetJSONMetric(t *testing.T) {
	type result struct {
		code int
		body schema.MetricReq
		hash string
	}

	tt := []struct {
		name      string
		req       schema.MetricReq
		serverKey security.Secret
		expected  result
	}{
		{
			name: "Get counter",
			req:  schema.NewGetCounterReq("PollCount"),
			expected: result{
				code: http.StatusOK,
				body: schema.NewUpdateCounterReq("PollCount", 10),
			},
		},
		{
			name: "Get gauge",
			req:  schema.NewGetGaugeReq("Alloc"),
			expected: result{
				code: http.StatusOK,
				body: schema.NewUpdateGaugeReq("Alloc", 11.345),
			},
		},
		{
			name:      "Should get signed counter",
			req:       schema.NewGetCounterReq("PollCount"),
			serverKey: "abc",
			expected: result{
				code: http.StatusOK,
				body: schema.NewUpdateCounterReq("PollCount", 10),
				hash: "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
			},
		},
		{
			name:      "Should get signed gauge",
			req:       schema.NewGetGaugeReq("Alloc"),
			serverKey: "abc",
			expected: result{
				code: http.StatusOK,
				body: schema.NewUpdateGaugeReq("Alloc", 11.345),
				hash: "2d32037265fd3547d65d4f51d69d8ea53490bef6e924fa2cfe2e4045ad50527d",
			},
		},
		{
			name: "Get unknown metric kind",
			req:  schema.MetricReq{ID: "Alloc", MType: "unknown"},
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name: "Get unknown counter",
			req:  schema.NewGetCounterReq("unknown"),
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get unknown gauge",
			req:  schema.NewGetGaugeReq("unknown"),
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Get counter with invalid name",
			req:  schema.NewGetCounterReq("X)"),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get gauge with invalid name",
			req:  schema.NewGetGaugeReq("X;"),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			payload, err := json.Marshal(tc.req)
			require.NoError(err)

			resp := sendTestRequest(t, http.MethodPost, "/value", payload, tc.serverKey)

			assert.Equal(tc.expected.code, resp.StatusCode)

			if tc.expected.code == http.StatusOK {
				assert.Equal("application/json", resp.Header.Get("Content-Type"))

				respBody, err := io.ReadAll(resp.Body)
				require.NoError(err)
				defer resp.Body.Close()

				var resp schema.MetricReq
				err = json.Unmarshal(respBody, &resp)
				require.NoError(err)

				tc.expected.body.Hash = tc.expected.hash
				assert.Equal(tc.expected.body, resp)
			}
		})
	}
}

func TestListMetrics(t *testing.T) {
	require := require.New(t)

	resp := sendTestRequest(t, http.MethodGet, "/", nil, "")

	require.Equal(http.StatusOK, resp.StatusCode)
	require.Equal("text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	require.NoError(err)
	require.NotZero(len(respBody))
}
