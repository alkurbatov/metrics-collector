package monitoring

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"golang.org/x/sync/errgroup"
)

// Metrics represent set of metrics collected by the agent.
type Metrics struct {
	// System metrics.
	System SystemStats

	// Metrics of Go runtime.
	Runtime RuntimeStats

	// Some random value required by Praktikum tasks.
	RandomValue metrics.Gauge

	// Count of metrics polling attempts.
	PollCount metrics.Counter
}

// Poll refreshes values of metrics and increments PollCount.
func (m *Metrics) Poll(ctx context.Context) error {
	m.PollCount++
	m.RandomValue = metrics.Gauge(rand.Float64()) //nolint: gosec

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		m.Runtime.Poll()
		return nil
	})

	g.Go(func() error {
		if err := m.System.Poll(ctx); err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to gather metrics: %w", err)
	}

	return nil
}
