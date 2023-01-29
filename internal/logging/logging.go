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

func RequestsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		id := uuid.NewV4().String()

		logger := log.With().Str("req-id", id).Logger()
		ctx := logger.WithContext(r.Context())

		logger.Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("")

		next.ServeHTTP(ww, r.WithContext(ctx))

		logger.Info().
			Int("status", ww.Status()).
			Msg("")
	})
}
