package compression

import (
	"compress/gzip"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/logging"
)

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Content-Encoding")
		if len(encoding) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		logger := logging.GetLogger(r.Context())
		logger.Debug().Msg("Got request compressed with " + encoding)

		if !isGzipEncoded(encoding) {
			err := entity.EncodingNotSupportedError(encoding)

			logger.Error().Err(err).Msg("")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			logger.Error().Err(err).Msg("")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer reader.Close()
		r.Body = reader

		next.ServeHTTP(w, r)
	})
}
