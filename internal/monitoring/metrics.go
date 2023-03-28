// Package monitoring provides means to collect different types of metrics.
// The collected metrics should be exported later e.g. using the package exporter.
package monitoring

import (
	"context"
	"fmt"
	"math/rand"
	"time"

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

	// Pseudo-random generator to fill the RandomValue metric.
	generator *rand.Rand
}

func NewMetrics() *Metrics {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint: gosec
	return &Metrics{generator: r}
}

// Poll refreshes values of metrics and increments PollCount.
func (m *Metrics) Poll(ctx context.Context) error {
	m.PollCount++
	m.RandomValue = metrics.Gauge(m.generator.Float64())

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
