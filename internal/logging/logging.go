// Package logging implements basic logging routine.
package logging

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Setup initializes new logger to be used across packages of the project.
func Setup(debug bool) {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	output.TimeFormat = time.RFC822

	l := zerolog.New(output).
		With().
		Timestamp()

	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		l = l.Caller()
	}

	log.Logger = l.Logger()
}

func generateRequestID() string {
	return uuid.NewV4().String()
}

// RequestsLogger is net/http middleware which logs incoming requests and responses.
func RequestsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		id := generateRequestID()

		logger := log.With().Str("req-id", id).Logger()
		ctx := logger.WithContext(r.Context())

		l := logger.Info().
			Str("transport", entity.TransportHTTP).
			Str("method", r.Method).
			Str("url", r.URL.String())

		clientIP := r.Header.Get("X-Real-IP")
		if len(clientIP) != 0 {
			l.Str("client-ip", clientIP)
		}

		l.Msg("")

		next.ServeHTTP(ww, r.WithContext(ctx))

		logger.Info().
			Int("status", ww.Status()).
			Msg("")
	})
}

// UnaryRequestsInterceptor is grpc unary interceptor which logs incoming requests and responses.
func UnaryRequestsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	id := generateRequestID()

	logger := log.With().Str("req-id", id).Logger()
	ctx = logger.WithContext(ctx)

	l := logger.Info().
		Str("transport", entity.TransportGRPC).
		Str("method", info.FullMethod)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("x-real-ip")
		if len(values) > 0 {
			l.Str("client-ip", values[0])
		}
	}

	l.Msg("")

	resp, err := handler(ctx, req)

	status, ok := status.FromError(err)
	if ok {
		logger.Info().
			Str("status", status.Code().String()).
			Msg("")
	} else {
		logger.Info().
			Err(err).
			Msg("")
	}

	return resp, err
}
