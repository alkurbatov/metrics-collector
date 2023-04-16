package security_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alkurbatov/metrics-collector/internal/grpcbackend"
	"github.com/alkurbatov/metrics-collector/internal/security"
	"github.com/alkurbatov/metrics-collector/pkg/grpcapi"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

func requireEqualCode(t *testing.T, expected codes.Code, err error) {
	t.Helper()

	rv, ok := status.FromError(err)

	require.True(t, ok)
	require.Equal(t, expected, rv.Code())
}

func sendHTTPRequest(t *testing.T, clientIP, trustedSubnet string) int {
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

func sendGRPCRequest(
	t *testing.T,
	mockAPI *grpcbackend.MetricsServerMock,
	trustedSubnet string,
) (grpcapi.MetricsClient, func()) {
	t.Helper()
	require := require.New(t)

	_, subnet, err := net.ParseCIDR(trustedSubnet)
	require.NoError(err)

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(security.UnaryRequestsFilter(subnet)),
	)

	grpcapi.RegisterMetricsServer(srv, mockAPI)

	go func() {
		require.NoError(srv.Serve(lis))
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.Dial("", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(err)

	closer := func() {
		require.NoError(conn.Close())
		srv.Stop()
		require.NoError(lis.Close())
	}

	client := grpcapi.NewMetricsClient(conn)

	return client, closer
}

func TestFilterHTTPRequest(t *testing.T) {
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
			status := sendHTTPRequest(t, tc.clientIP, tc.trustedSubnet)

			require.Equal(t, tc.expected, status)
		})
	}
}

func TestFilterGRPCRequest(t *testing.T) {
	tt := []struct {
		name          string
		clientIP      string
		trustedSubnet string
		expected      codes.Code
	}{
		{
			name:          "Accepts IPv4 from trusted subnet",
			clientIP:      "192.168.0.10",
			trustedSubnet: "192.168.0.0/16",
			expected:      codes.OK,
		},
		{
			name:          "Accepts IPv6 from trusted subnet",
			clientIP:      "::1",
			trustedSubnet: "::1/128",
			expected:      codes.OK,
		},
		{
			name:          "Rejects IPv4 not from trusted subnet",
			clientIP:      "10.30.10.12",
			trustedSubnet: "192.168.0.0/32",
			expected:      codes.PermissionDenied,
		},
		{
			name:          "Rejects IPv6 not from trusted subnet",
			clientIP:      "2001:470:28:30d:21e:67ff:fe7a:e1da",
			trustedSubnet: "::1/128",
			expected:      codes.PermissionDenied,
		},
		{
			name:          "Rejects not set IP",
			trustedSubnet: "192.168.0.0/32",
			expected:      codes.PermissionDenied,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := new(grpcbackend.MetricsServerMock)
			m.On("Update", mock.Anything, mock.AnythingOfType("*grpcapi.MetricReq")).Return(&grpcapi.MetricReq{}, nil)

			client, closer := sendGRPCRequest(t, m, tc.trustedSubnet)
			t.Cleanup(closer)

			md := metadata.New(map[string]string{"x-real-ip": tc.clientIP})
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			_, err := client.Update(ctx, &grpcapi.MetricReq{})

			requireEqualCode(t, tc.expected, err)
		})
	}
}

func TestFilterGRPCRequestAppliedToBatchUpdate(t *testing.T) {
	m := new(grpcbackend.MetricsServerMock)
	m.On("BatchUpdate", mock.Anything, mock.AnythingOfType("*grpcapi.BatchUpdateRequest")).
		Return(&emptypb.Empty{}, nil)

	client, closer := sendGRPCRequest(t, m, "192.168.0.0/32")
	t.Cleanup(closer)

	_, err := client.BatchUpdate(context.Background(), &grpcapi.BatchUpdateRequest{})
	requireEqualCode(t, codes.PermissionDenied, err)
}

func TestFilterGRPCRequestNotAppliedToGet(t *testing.T) {
	m := new(grpcbackend.MetricsServerMock)
	m.On("Get", mock.Anything, mock.AnythingOfType("*grpcapi.GetMetricRequest")).
		Return(&grpcapi.MetricReq{}, nil)

	client, closer := sendGRPCRequest(t, m, "192.168.0.0/32")
	t.Cleanup(closer)

	_, err := client.Get(context.Background(), &grpcapi.GetMetricRequest{})
	requireEqualCode(t, codes.OK, err)
}
