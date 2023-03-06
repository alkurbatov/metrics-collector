package monitoring

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"golang.org/x/sync/errgroup"
)

type Metrics struct {
	Process ProcessStats
	Runtime RuntimeStats

	RandomValue metrics.Gauge

	PollCount metrics.Counter
}

func (m *Metrics) Poll(ctx context.Context) error {
	m.PollCount++
	m.RandomValue = metrics.Gauge(rand.Float64()) //nolint: gosec

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		m.Runtime.Poll()
		return nil
	})

	g.Go(func() error {
		if err := m.Process.Poll(ctx); err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to gather metrics: %w", err)
	}

	return nil
}
