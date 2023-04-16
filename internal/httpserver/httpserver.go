// Package httpserver implements handy wrap around HTTP server
// to group common settings and tasks inside single entity.
package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

// Set reasonable timeouts, see:
// https://habr.com/ru/company/ispring/blog/560032/
const (
	_defaultReadTimeout       = 5 * time.Second
	_defaultWriteTimeout      = 10 * time.Second
	_defaultIdleTimeout       = 120 * time.Second
	_defaultReadHeaderTimeout = 5 * time.Second
)

// Server wraps HTTP server entity and handy means
// to simplify work with the entity.
type Server struct {
	server *http.Server
	notify chan error
}

// New creates and initializes new instance of HTTP server.
func New(handler http.Handler, address entity.NetAddress) *Server {
	httpServer := &http.Server{
		Handler:           handler,
		Addr:              address.String(),
		ReadTimeout:       _defaultReadTimeout,
		WriteTimeout:      _defaultWriteTimeout,
		IdleTimeout:       _defaultIdleTimeout,
		ReadHeaderTimeout: _defaultReadHeaderTimeout,
	}

	s := &Server{
		server: httpServer,
		notify: make(chan error, 1),
	}

	return s
}

// Start launches the HTTP server.
func (s *Server) Start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

// Notify reports errors received during start and work of the server.
// Usually such errors are not recoverable.
func (s *Server) Notify() <-chan error {
	return s.notify
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	return s.server.Shutdown(ctx)
}
