package security

import (
	"net"
	"net/http"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/rs/zerolog/log"
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
