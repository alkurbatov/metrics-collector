package httpbackend_test

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/httpbackend"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/internal/services"
	"github.com/alkurbatov/metrics-collector/internal/storage"
	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRouter(
	t *testing.T,
	key security.Secret,
	recorder services.Recorder,
	healthcheck services.HealthCheck,
) http.Handler {
	t.Helper()

	var signer *security.Signer
	if len(key) > 0 {
		signer = security.NewSigner(key)
	}

	view, err := template.ParseFiles("../../web/views/metrics.html")
	require.NoError(t, err)

	return httpbackend.Router("0.0.0.0:8080", view, recorder, healthcheck, signer, nil, nil)
}

func sendTestRequest(t *testing.T, router http.Handler, method, path string, payload []byte) (int, string, []byte) {
	t.Helper()

	srv := httptest.NewServer(router)
	defer srv.Close()

	body := bytes.NewReader(payload)

	req, err := http.NewRequest(method, srv.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() {
		_ = resp.Body.Close()
	}()

	contentType := resp.Header.Get("Content-Type")

	if resp.Body == http.NoBody {
		return resp.StatusCode, contentType, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, contentType, respBody
}

func TestUpdateMetric(t *testing.T) {
	type result struct {
		code int
		body string
	}

	tt := []struct {
		name        string
		path        string
		recorderRV  storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should push counter",
			path:       "/update/counter/PollCount/10",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(20)},
			expected: result{
				code: http.StatusOK,
				body: "20",
			},
		},
		{
			name:       "Should push gauge",
			path:       "/update/gauge/Alloc/13.123",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected: result{
				code: http.StatusOK,
				body: "13.123",
			},
		},
		{
			name: "Should fail on unknown metric kind",
			path: "/update/unknown/Alloc/12.123",
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name: "Should fail on counter with invalid name",
			path: "/update/counter/X)/10",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail on gauge with invalid name",
			path: "/update/gauge/X;/10.234",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail on counter with invalid value",
			path: "/update/counter/PollCount/10\\.0",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail gauge with invalid value",
			path: "/update/gauge/Alloc/15.234;",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "Should fail on broken recorder",
			path:        "/update/counter/fail/10",
			recorderErr: entity.ErrUnexpected,
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			m := new(services.RecorderMock)
			m.On("Push", mock.Anything, mock.AnythingOfType("Record")).Return(tc.recorderRV, tc.recorderErr)

			router := newRouter(t, "", m, nil)
			code, _, body := sendTestRequest(t, router, http.MethodPost, tc.path, nil)

			assert.Equal(tc.expected.code, code)

			if tc.expected.code == http.StatusOK {
				assert.Equal(tc.expected.body, string(body))
			}
		})
	}
}

func TestUpdateJSONMetric(t *testing.T) {
	type result struct {
		code int
	}

	tt := []struct {
		name        string
		req         metrics.MetricReq
		clientKey   security.Secret
		serverKey   security.Secret
		recorderRV  storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should push counter",
			req:        metrics.NewUpdateCounterReq("PollCount", 10),
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:       "Should push gauge",
			req:        metrics.NewUpdateGaugeReq("Alloc", 13.123),
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:       "Should push counter with signature",
			req:        metrics.NewUpdateCounterReq("PollCount", 10),
			clientKey:  "abc",
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:       "Should push gauge with signature",
			req:        metrics.NewUpdateGaugeReq("Alloc", 13.123),
			clientKey:  "abc",
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:       "Should fail if counter signature doesn't match",
			req:        metrics.NewUpdateCounterReq("PollCount", 10),
			clientKey:  "abc",
			serverKey:  "xxx",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:       "Should fail if gauge signature doesn't match",
			req:        metrics.NewUpdateGaugeReq("Alloc", 13.123),
			clientKey:  "abc",
			serverKey:  "xxx",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(13.123)},
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail on unknown metric kind",
			req: metrics.MetricReq{
				ID:    "X",
				MType: "unknown",
			},
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name: "Should fail on counter with invalid name",
			req:  metrics.NewUpdateCounterReq("X)", 10),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail on gauge with invalid name",
			req:  metrics.NewUpdateGaugeReq("X;", 13.123),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "Should fail on broken recorder",
			req:         metrics.NewUpdateCounterReq("fail", 13),
			recorderErr: entity.ErrUnexpected,
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m := new(services.RecorderMock)
			m.On("Push", mock.Anything, mock.AnythingOfType("Record")).Return(tc.recorderRV, tc.recorderErr)

			router := newRouter(t, tc.serverKey, m, nil)

			if len(tc.clientKey) > 0 {
				signer := security.NewSigner(tc.clientKey)
				hash, err := signer.CalculateRecordSignature(tc.recorderRV)
				require.NoError(err)
				tc.req.Hash = hash
			}

			payload, err := json.Marshal(tc.req)
			require.NoError(err)

			code, contentType, body := sendTestRequest(t, router, http.MethodPost, "/update", payload)

			assert.Equal(tc.expected.code, code)

			if tc.expected.code == http.StatusOK {
				assert.Equal("application/json", contentType)

				var resp metrics.MetricReq
				err = json.Unmarshal(body, &resp)
				require.NoError(err)

				assert.Equal(tc.req, resp)
			}
		})
	}
}

func TestBatchUpdate(t *testing.T) {
	type result struct {
		code int
	}

	batchReq := []metrics.MetricReq{
		metrics.NewUpdateCounterReq("PollCount", 10),
		metrics.NewUpdateGaugeReq("Alloc", 11.23),
	}

	batchResp := []storage.Record{
		{Name: "PollCount", Value: metrics.Counter(10)},
		{Name: "Alloc", Value: metrics.Gauge(11.23)},
	}

	tt := []struct {
		name        string
		req         []metrics.MetricReq
		clientKey   security.Secret
		serverKey   security.Secret
		recorderRv  []storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should handle signed list of different metrics",
			req:        batchReq,
			clientKey:  "abc",
			serverKey:  "abc",
			recorderRv: batchResp,
			expected:   result{code: http.StatusOK},
		},
		{
			name:       "Should handle unsigned list of different metrics",
			req:        batchReq,
			recorderRv: batchResp,
			expected:   result{code: http.StatusOK},
		},
		{
			name:     "Should fail on empty list",
			req:      make([]metrics.MetricReq, 0),
			expected: result{code: http.StatusBadRequest},
		},
		{
			name:      "Should fail on wrong signature",
			req:       batchReq,
			clientKey: "abc",
			serverKey: "xxx",
			expected:  result{code: http.StatusBadRequest},
		},
		{
			name:     "Should fail if counter value is missing",
			req:      []metrics.MetricReq{{ID: "xxx", MType: "counter"}},
			expected: result{code: http.StatusBadRequest},
		},
		{
			name:     "Should fail if gauge value missing",
			req:      []metrics.MetricReq{{ID: "xxx", MType: "gauge"}},
			expected: result{code: http.StatusBadRequest},
		},
		{
			name:     "Should fail in unknown metric kind found in list",
			req:      []metrics.MetricReq{{ID: "xxx", MType: "unknown"}},
			expected: result{code: http.StatusNotImplemented},
		},
		{
			name:        "Should fail if recorder is broken",
			req:         batchReq,
			recorderRv:  nil,
			recorderErr: entity.ErrUnexpected,
			expected:    result{code: http.StatusInternalServerError},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			m := new(services.RecorderMock)
			m.On("PushList", mock.Anything, mock.Anything).Return(tc.recorderRv, tc.recorderErr)

			router := newRouter(t, tc.serverKey, m, nil)

			if len(tc.clientKey) > 0 {
				signer := security.NewSigner(tc.clientKey)

				hash, err := signer.CalculateSignature(tc.req[0].ID, *tc.req[0].Delta)
				require.NoError(err)
				tc.req[0].Hash = hash

				hash, err = signer.CalculateSignature(tc.req[1].ID, *tc.req[1].Value)
				require.NoError(err)
				tc.req[1].Hash = hash
			}

			payload, err := json.Marshal(tc.req)
			require.NoError(err)

			code, _, _ := sendTestRequest(t, router, http.MethodPost, "/updates/", payload)
			require.Equal(tc.expected.code, code)
		})
	}
}

func TestGetMetric(t *testing.T) {
	type result struct {
		code int
		body string
	}

	tt := []struct {
		name        string
		path        string
		recorderRV  storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should get counter",
			path:       "/value/counter/PollCount",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: http.StatusOK,
				body: "10",
			},
		},
		{
			name:       "Should get gauge",
			path:       "/value/gauge/Alloc",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: http.StatusOK,
				body: "11.345",
			},
		},
		{
			name: "Should fail if metric kind unknown",
			path: "/value/unknown/Alloc",
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name:        "Should fail on unknown counter",
			path:        "/value/counter/unknown",
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "Should fail on unknown gauge",
			path:        "/value/gauge/unknown",
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Should fail on counter with invalid name",
			path: "/value/counter/X)",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail on gauge with invalid name",
			path: "/value/gauge/X;",
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "Should fail on broken recorder",
			path:        "/value/gauge/Alloc",
			recorderErr: entity.ErrUnexpected,
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			m := new(services.RecorderMock)
			m.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(tc.recorderRV, tc.recorderErr)

			router := newRouter(t, "", m, nil)

			code, contentType, body := sendTestRequest(t, router, http.MethodGet, tc.path, nil)

			assert.Equal(tc.expected.code, code)
			assert.Equal("text/plain; charset=utf-8", contentType)

			if tc.expected.code == http.StatusOK {
				assert.Equal(tc.expected.body, string(body))
			}
		})
	}
}

func TestGetJSONMetric(t *testing.T) {
	type result struct {
		code int
		body metrics.MetricReq
		hash string
	}

	tt := []struct {
		name        string
		req         metrics.MetricReq
		serverKey   security.Secret
		recorderRV  storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should get counter",
			req:        metrics.NewGetCounterReq("PollCount"),
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: http.StatusOK,
				body: metrics.NewUpdateCounterReq("PollCount", 10),
			},
		},
		{
			name:       "Should get gauge",
			req:        metrics.NewGetGaugeReq("Alloc"),
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: http.StatusOK,
				body: metrics.NewUpdateGaugeReq("Alloc", 11.345),
			},
		},
		{
			name:       "Should get signed counter",
			req:        metrics.NewGetCounterReq("PollCount"),
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "PollCount", Value: metrics.Counter(10)},
			expected: result{
				code: http.StatusOK,
				body: metrics.NewUpdateCounterReq("PollCount", 10),
				hash: "0833001195f2e062140968e0c00dd44f00eb9a0b309aedc464817f904b244c8a",
			},
		},
		{
			name:       "Should get signed gauge",
			req:        metrics.NewGetGaugeReq("Alloc"),
			serverKey:  "abc",
			recorderRV: storage.Record{Name: "Alloc", Value: metrics.Gauge(11.345)},
			expected: result{
				code: http.StatusOK,
				body: metrics.NewUpdateGaugeReq("Alloc", 11.345),
				hash: "2d32037265fd3547d65d4f51d69d8ea53490bef6e924fa2cfe2e4045ad50527d",
			},
		},
		{
			name: "Should fail on unknown metric kind",
			req:  metrics.MetricReq{ID: "Alloc", MType: "unknown"},
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name:        "Should fail on unknown counter",
			req:         metrics.NewGetCounterReq("unknown"),
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "Should fail on unknown gauge",
			req:         metrics.NewGetGaugeReq("unknown"),
			recorderErr: entity.ErrMetricNotFound,
			expected: result{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Should fail on counter with invalid name",
			req:  metrics.NewGetCounterReq("X)"),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Should fail on gauge with invalid name",
			req:  metrics.NewGetGaugeReq("X;"),
			expected: result{
				code: http.StatusBadRequest,
			},
		},
		{
			name:        "Should fail on broken recorder",
			req:         metrics.NewGetGaugeReq("X"),
			recorderErr: entity.ErrUnexpected,
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			m := new(services.RecorderMock)
			m.On("Get", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(tc.recorderRV, tc.recorderErr)

			router := newRouter(t, tc.serverKey, m, nil)

			payload, err := json.Marshal(tc.req)
			require.NoError(err)

			code, contentType, body := sendTestRequest(t, router, http.MethodPost, "/value", payload)
			assert.Equal(tc.expected.code, code)

			if tc.expected.code == http.StatusOK {
				assert.Equal("application/json", contentType)

				var resp metrics.MetricReq
				err = json.Unmarshal(body, &resp)
				require.NoError(err)

				tc.expected.body.Hash = tc.expected.hash
				assert.Equal(tc.expected.body, resp)
			}
		})
	}
}

func TestListMetrics(t *testing.T) {
	stored := []storage.Record{
		{Name: "A", Value: metrics.Counter(10)},
		{Name: "B", Value: metrics.Gauge(11.345)}}

	type result struct {
		code int
	}

	tt := []struct {
		name        string
		recorderRV  []storage.Record
		recorderErr error
		expected    result
	}{
		{
			name:       "Should provide HTML page with all metrics",
			recorderRV: stored,
			expected:   result{code: http.StatusOK},
		},
		{
			name:        "Should fail on broken recorder",
			recorderErr: entity.ErrUnexpected,
			expected:    result{code: http.StatusInternalServerError},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(services.RecorderMock)
			m.On("List", mock.Anything).Return(tc.recorderRV, tc.recorderErr)

			router := newRouter(t, "", m, nil)
			require := require.New(t)

			code, contentType, body := sendTestRequest(t, router, http.MethodGet, "/", nil)

			require.Equal(tc.expected.code, code)

			if tc.expected.code == http.StatusOK {
				require.Equal("text/html; charset=utf-8", contentType)
			}

			require.NotZero(len(body))
		})
	}
}

func TestPing(t *testing.T) {
	type result struct {
		code int
	}

	tt := []struct {
		name      string
		checkResp error
		expected  result
	}{
		{
			name:      "Should return OK, if storage online",
			checkResp: nil,
			expected: result{
				code: http.StatusOK,
			},
		},
		{
			name:      "Should return not implemented, if storage doesn't support health check",
			checkResp: entity.ErrHealthCheckNotSupported,
			expected: result{
				code: http.StatusNotImplemented,
			},
		},
		{
			name:      "Should return internal error, if storage offline",
			checkResp: entity.ErrUnexpected,
			expected: result{
				code: http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require := require.New(t)

			m := new(services.HealthCheckMock)
			m.On("CheckStorage", mock.Anything).Return(tc.checkResp)

			router := newRouter(t, "", nil, m)

			code, _, body := sendTestRequest(t, router, http.MethodGet, "/ping", nil)

			require.Equal(tc.expected.code, code)

			if tc.expected.code == http.StatusOK {
				require.Nil(body)
			} else {
				require.NotZero(len(body))
			}
		})
	}
}
