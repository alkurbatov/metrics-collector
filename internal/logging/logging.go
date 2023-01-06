package logging

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

type loggerKey string

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
		ctx := context.WithValue(r.Context(), loggerKey("logger"), &logger)

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

func GetLogger(ctx context.Context) *zerolog.Logger {
	if v := ctx.Value(loggerKey("logger")); v != nil {
		return v.(*zerolog.Logger)
	}

	return &log.Logger
}
