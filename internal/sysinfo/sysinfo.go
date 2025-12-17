package sysinfo

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/tshojoshua/jtnt-agent/pkg/api"
)

// Collector collects system information
type Collector struct{}

// NewCollector creates a new system info collector
func NewCollector() *Collector {
	return &Collector{}
}

// Collect gathers current system information
func (c *Collector) Collect() (*api.SystemInfo, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Get OS info
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	// Get CPU info
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %w", err)
	}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	var cpuUsage float64
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Get disk info
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get disk info: %w", err)
	}

	// Get IP addresses
	ipAddrs, err := c.getIPAddresses()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP addresses: %w", err)
	}

	return &api.SystemInfo{
		Hostname:    hostname,
		OS:          hostInfo.OS,
		OSVersion:   hostInfo.PlatformVersion,
		Arch:        hostInfo.KernelArch,
		Uptime:      int64(hostInfo.Uptime),
		CPUCount:    cpuCount,
		CPUUsage:    cpuUsage,
		MemTotal:    memInfo.Total,
		MemUsed:     memInfo.Used,
		DiskTotal:   diskInfo.Total,
		DiskUsed:    diskInfo.Used,
		IPAddresses: ipAddrs,
		Timestamp:   time.Now(),
	}, nil
}

func (c *Collector) getIPAddresses() ([]string, error) {
	var ips []string

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Only include IPv4 for now
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}

	return ips, nil
}
