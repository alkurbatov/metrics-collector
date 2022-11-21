package middleware

import (
	"fmt"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

func RequestsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.Log.Info(fmt.Sprintf("%s %s", r.Method, r.URL.String()))

		next.ServeHTTP(w, r)
	})
}
