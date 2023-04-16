package security

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// FilterRequest is a HTTP middleware that rejects requests which
// don't match trusted subnet.
func FilterRequest(trustedSubnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			clientIP := net.ParseIP(r.Header.Get("X-Real-IP"))

			if !trustedSubnet.Contains(clientIP) {
				logger := log.Ctx(r.Context())
				logger.Error().Err(entity.UntrustedSourceError(clientIP)).Msg("security - FilterRequest - trustedSubnet.Contains")
				http.Error(w, "", http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

// UnaryRequestsFilter is grpc unary interceptor that rejects requests which
// don't match trusted subnet.
func UnaryRequestsFilter(
	trustedSubnet *net.IPNet,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if !strings.HasSuffix(info.FullMethod, "Update") {
			return handler(ctx, req)
		}

		var clientIP net.IP

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				clientIP = net.ParseIP(values[0])
			}
		}

		if !trustedSubnet.Contains(clientIP) {
			err := entity.UntrustedSourceError(clientIP)

			logger := log.Ctx(ctx)
			logger.Error().Err(err).Msg("security - UnaryRequestsFilter - trustedSubnet.Contains")

			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		return handler(ctx, req)
	}
}
