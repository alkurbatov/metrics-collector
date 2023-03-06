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

type ProcessStats struct {
	TotalMemory     metrics.Gauge
	FreeMemory      metrics.Gauge
	CPUutilization1 metrics.Gauge // utilisation in percent of all available logical/virtual cores
}

func (p *ProcessStats) Poll(ctx context.Context) error {
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return processPollError(err)
	}

	p.TotalMemory = metrics.Gauge(vMem.Total)
	p.FreeMemory = metrics.Gauge(vMem.Free)

	utilisation, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return processPollError(err)
	}

	// NB (alkurbatov): Since we asked for cumulative CPU stats
	// there is only single value in the array.
	p.CPUutilization1 = metrics.Gauge(utilisation[0])

	return nil
}
