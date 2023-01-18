package metrics

import (
	"context"
	"fmt"
	"math/rand"

	"golang.org/x/sync/errgroup"
)

type Metrics struct {
	Process ProcessStats
	Runtime RuntimeStats

	RandomValue Gauge

	PollCount Counter
}

func (m *Metrics) Poll(ctx context.Context) error {
	m.PollCount++
	m.RandomValue = Gauge(rand.Float64()) //nolint: gosec

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
