package logging

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
}

func RequestsLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Info(fmt.Sprintf("%s %s", r.Method, r.URL.String()))

		next.ServeHTTP(w, r)
	})
}
