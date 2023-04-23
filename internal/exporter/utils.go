package exporter

import (
	"fmt"
	"net"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

// Get preferred outbound IP address of this machine.
func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IP{}, err
	}

	defer func() {
		_ = conn.Close()
	}()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return net.IP{}, fmt.Errorf("exporter - getOutboundIP - conn.LocalAddr: %w", entity.ErrUnexpected)
	}

	return localAddr.IP, nil
}
