package sysinfo

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemInfo struct {
	Hostname    string    `json:"hostname"`
	OS          string    `json:"os"`
	OSVersion   string    `json:"os_version"`
	Arch        string    `json:"arch"`
	Uptime      uint64    `json:"uptime"`
	CPUCount    int       `json:"cpu_count"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemTotal    uint64    `json:"mem_total"`
	MemUsed     uint64    `json:"mem_used"`
	MemPercent  float64   `json:"mem_percent"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskUsed    uint64    `json:"disk_used"`
	DiskPercent float64   `json:"disk_percent"`
	IPAddresses []string  `json:"ip_addresses"`
	Timestamp   time.Time `json:"timestamp"`
}

func Collect() (*SystemInfo, error) {
	info := &SystemInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Timestamp: time.Now(),
	}

	// Hostname
	if hostname, err := host.Info(); err == nil {
		info.Hostname = hostname.Hostname
		info.OSVersion = hostname.Platform + " " + hostname.PlatformVersion
		info.Uptime = hostname.Uptime
	}

	// CPU
	info.CPUCount = runtime.NumCPU()
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		info.CPUUsage = cpuPercent[0]
	}

	// Memory
	if vmem, err := mem.VirtualMemory(); err == nil {
		info.MemTotal = vmem.Total
		info.MemUsed = vmem.Used
		info.MemPercent = vmem.UsedPercent
	}

	// Disk (root partition)
	if usage, err := disk.Usage(getRootPath()); err == nil {
		info.DiskTotal = usage.Total
		info.DiskUsed = usage.Used
		info.DiskPercent = usage.UsedPercent
	}

	// Network interfaces
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			for _, addr := range iface.Addrs {
				info.IPAddresses = append(info.IPAddresses, addr.Addr)
			}
		}
	}

	return info, nil
}

func getRootPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	}
	return "/"
}
