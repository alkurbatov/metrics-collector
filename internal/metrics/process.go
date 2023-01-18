package metrics

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func processPollError(reason error) error {
	return fmt.Errorf("process poll failed: %w", reason)
}

type ProcessStats struct {
	TotalMemory     Gauge
	FreeMemory      Gauge
	CPUutilization1 Gauge // utilisation in percent of all available logical/virtual cores
}

func (p *ProcessStats) Poll(ctx context.Context) error {
	vMem, err := mem.VirtualMemory()
	if err != nil {
		return processPollError(err)
	}

	p.TotalMemory = Gauge(vMem.Total)
	p.FreeMemory = Gauge(vMem.Free)

	utilisation, err := cpu.PercentWithContext(ctx, 0, false)
	if err != nil {
		return processPollError(err)
	}

	// NB (alkurbatov): Since we asked for cumulative CPU stats
	// there is only single value in the array.
	p.CPUutilization1 = Gauge(utilisation[0])

	return nil
}
