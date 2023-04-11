package security_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

func sendRequest(t *testing.T, clientIP, trustedSubnet string) int {
	t.Helper()
	require := require.New(t)

	_, subnet, err := net.ParseCIDR(trustedSubnet)
	require.NoError(err)

	router := chi.NewRouter()
	router.Use(security.FilterRequest(subnet))
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, nil)
	require.NoError(err)

	req.Header.Set("X-Real-IP", clientIP)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(err)

	defer func() {
		_ = resp.Body.Close()
	}()

	return resp.StatusCode
}

func TestFilterRequest(t *testing.T) {
	tt := []struct {
		name          string
		clientIP      string
		trustedSubnet string
		expected      int
	}{
		{
			name:          "Accepts IPv4 from trusted subnet",
			clientIP:      "192.168.0.10",
			trustedSubnet: "192.168.0.0/16",
			expected:      http.StatusOK,
		},
		{
			name:          "Accepts IPv6 from trusted subnet",
			clientIP:      "::1",
			trustedSubnet: "::1/128",
			expected:      http.StatusOK,
		},
		{
			name:          "Rejects IPv4 not from trusted subnet",
			clientIP:      "10.30.10.12",
			trustedSubnet: "192.168.0.0/32",
			expected:      http.StatusForbidden,
		},
		{
			name:          "Rejects IPv6 not from trusted subnet",
			clientIP:      "2001:470:28:30d:21e:67ff:fe7a:e1da",
			trustedSubnet: "::1/128",
			expected:      http.StatusForbidden,
		},
		{
			name:          "Rejects not set IP",
			trustedSubnet: "192.168.0.0/32",
			expected:      http.StatusForbidden,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			status := sendRequest(t, tc.clientIP, tc.trustedSubnet)

			require.Equal(t, tc.expected, status)
		})
	}
}
