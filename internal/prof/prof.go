// Package prof encapsulates pprof with attached HTTP server.
package prof

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/logging"
)

// Profiler is a separate HTTP server serving pprof requests.
type Profiler struct {
	server *http.Server
}

// New creates Profiler instance listening on specified address and port.
// To start the underlying HTTP server one should call the Start function.
func New(address entity.NetAddress) *Profiler {
	r := chi.NewRouter()

	r.Use(logging.RequestsLogger)
	r.Mount("/debug", middleware.Profiler())

	httpServer := &http.Server{
		Handler:     r,
		Addr:        address.String(),
		ReadTimeout: 5 * time.Second,
	}

	p := &Profiler{
		server: httpServer,
	}

	// NB (alkurbatov): Tweak memory profiling rate to spot more allocations.
	runtime.MemProfileRate = 2048

	return p
}

// Start runs the HTTP server handling pprof requests.
func (p *Profiler) Start() {
	if err := p.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("Profiler - Start - p.server.ListenAndServe")
	}
}

// Shutdown stops the underlying HTTP server.
func (p *Profiler) Shutdown(ctx context.Context) error {
	if err := p.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Profiler - Start - p.server.Shutdown: %w", err)
	}

	return nil
}
