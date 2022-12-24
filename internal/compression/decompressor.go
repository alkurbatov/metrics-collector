package compression

import (
	"compress/gzip"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Content-Encoding")
		if len(encoding) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		logging.Log.Debug("Got request compressed with " + encoding)

		if !isGzipEncoded(encoding) {
			err := "compression type " + encoding + " not supported"

			logging.Log.Error(err)
			http.Error(w, err, http.StatusBadRequest)
			return
		}

		reader, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer reader.Close()
		r.Body = reader

		next.ServeHTTP(w, r)
	})
}
