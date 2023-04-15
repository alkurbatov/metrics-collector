// Package prof encapsulates pprof with attached HTTP server.
package prof

import (
	"runtime"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/alkurbatov/metrics-collector/internal/entity"
	"github.com/alkurbatov/metrics-collector/internal/httpserver"
	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type Profiler struct {
	*httpserver.Server
}

// New creates Profiler instance listening on specified address and port.
// If empty address string used, empty Profiler (without initialization)
// is created. Such profiler is no-op.
func New(address entity.NetAddress) *Profiler {
	if len(address) == 0 {
		return &Profiler{&httpserver.Server{}}
	}

	r := chi.NewRouter()

	r.Use(logging.RequestsLogger)
	r.Mount("/debug", middleware.Profiler())

	// NB (alkurbatov): Tweak memory profiling rate to spot more allocations.
	runtime.MemProfileRate = 2048

	server := httpserver.New(r, address)
	server.Start()

	return &Profiler{server}
}
