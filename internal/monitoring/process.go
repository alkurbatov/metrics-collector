package monitoring

import (
	"context"
	"fmt"

	"github.com/alkurbatov/metrics-collector/pkg/metrics"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func processPollError(reason error) error {
	return fmt.Errorf("process poll failed: %w", reason)
}

type SystemStats struct {
	// Total amount of RAM on this system.
	TotalMemory metrics.Gauge

	// Total amount of not used RAM on this system
	FreeMemory metrics.Gauge

	// Utilisation in percent of all available logical/virtual cores on this system.
	CPUutilization1 metrics.Gauge
}

// Poll refreshes values of system metrics.
func (s *SystemStats) Poll(ctx context.Context) error {
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return processPollError(err)
	}

	s.TotalMemory = metrics.Gauge(vMem.Total)
	s.FreeMemory = metrics.Gauge(vMem.Free)

	utilisation, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return processPollError(err)
	}

	// NB (alkurbatov): Since we asked for cumulative CPU stats
	// there is only single value in the array.
	s.CPUutilization1 = metrics.Gauge(utilisation[0])

	return nil
}
