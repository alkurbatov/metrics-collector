package prof

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	_ "net/http/pprof" //nolint: gosec //served on different port which should be hidden in production

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type Profiler struct {
	server *http.Server
}

func New(address entity.NetAddress) *Profiler {
	r := chi.NewRouter()

	r.Use(logging.RequestsLogger)
	r.Use(middleware.StripSlashes)

	r.Mount("/debug/pprof", http.DefaultServeMux)

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

func (p *Profiler) Start() {
	if err := p.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("Profiler - Start - p.server.ListenAndServe")
	}
}

func (p *Profiler) Shutdown(ctx context.Context) error {
	if err := p.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("Profiler - Start - p.server.Shutdown: %w", err)
	}

	return nil
}
