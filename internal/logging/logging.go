// Package logging implements basic logging routine.
package logging

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
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

// RequestsLogger is net/http middleware used to log incoming requests and responses.
func RequestsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		id := uuid.NewV4().String()

		logger := log.With().Str("req-id", id).Logger()
		ctx := logger.WithContext(r.Context())

		l := logger.Info().
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
