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
// Returns partial information if some metrics fail (e.g., in containers)
func (c *Collector) Collect() (*api.SystemInfo, error) {
	sysInfo := &api.SystemInfo{
		Timestamp: time.Now(),
	}

	// Get hostname - this is critical
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	sysInfo.Hostname = hostname

	// Get OS info - best effort in containers
	hostInfo, err := host.Info()
	if err == nil {
		sysInfo.OS = hostInfo.OS
		sysInfo.OSVersion = hostInfo.PlatformVersion
		sysInfo.Arch = hostInfo.KernelArch
		sysInfo.Uptime = int64(hostInfo.Uptime)
	} else {
		// In containers, we may not have access to /proc/stat
		// Provide minimal fallback info
		sysInfo.OS = "unknown"
		sysInfo.OSVersion = "unknown"
		sysInfo.Arch = "unknown"
		sysInfo.Uptime = 0
	}

	// Get CPU info - best effort
	cpuCount, err := cpu.Counts(true)
	if err == nil {
		sysInfo.CPUCount = cpuCount
	}

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		sysInfo.CPUUsage = cpuPercent[0]
	}

	// Get memory info - best effort
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		sysInfo.MemTotal = memInfo.Total
		sysInfo.MemUsed = memInfo.Used
	}

	// Get disk info - best effort
	diskInfo, err := disk.Usage("/")
	if err == nil {
		sysInfo.DiskTotal = diskInfo.Total
		sysInfo.DiskUsed = diskInfo.Used
	}

	// Get IP addresses - best effort
	ipAddrs, err := c.getIPAddresses()
	if err == nil {
		sysInfo.IPAddresses = ipAddrs
	}

	return sysInfo, nil
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
