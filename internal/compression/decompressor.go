package compression

import (
	"compress/gzip"
	"net/http"
	"sync"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/rs/zerolog/log"
)

var gzipReadersPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Content-Encoding")
		if len(encoding) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		logger := log.Ctx(r.Context())
		logger.Debug().Msg("Got request compressed with " + encoding)

		if !isGzipEncoded(encoding) {
			err := entity.EncodingNotSupportedError(encoding)

			logger.Error().Err(err).Msg("")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reader := gzipReadersPool.Get().(*gzip.Reader)
		if err := reader.Reset(r.Body); err != nil {
			logger.Error().Err(err).Msg("DecompressRequest - reader.Reset")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer func() {
			if err := reader.Close(); err != nil {
				logger.Error().Err(err).Msg("DecompressRequest - reader.Close")
			}

			gzipReadersPool.Put(reader)
		}()

		r.Body = reader

		next.ServeHTTP(w, r)
	})
}
