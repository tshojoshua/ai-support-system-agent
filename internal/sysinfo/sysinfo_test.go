package sysinfo

import (
	"testing"
)

func TestCollector_Collect(t *testing.T) {
	collector := NewCollector()

	info, err := collector.Collect()
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}

	// Validate required fields
	if info.Hostname == "" {
		t.Error("Hostname should not be empty")
	}

	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	if info.Arch == "" {
		t.Error("Arch should not be empty")
	}

	if info.CPUCount <= 0 {
		t.Errorf("CPUCount should be positive, got %d", info.CPUCount)
	}

	if info.MemTotal == 0 {
		t.Error("MemTotal should not be zero")
	}

	if info.DiskTotal == 0 {
		t.Error("DiskTotal should not be zero")
	}

	// CPU usage should be between 0 and 100
	if info.CPUUsage < 0 || info.CPUUsage > 100 {
		t.Errorf("CPUUsage should be 0-100, got %.2f", info.CPUUsage)
	}

	// Memory used should not exceed total
	if info.MemUsed > info.MemTotal {
		t.Errorf("MemUsed (%d) should not exceed MemTotal (%d)", info.MemUsed, info.MemTotal)
	}

	// Disk used should not exceed total
	if info.DiskUsed > info.DiskTotal {
		t.Errorf("DiskUsed (%d) should not exceed DiskTotal (%d)", info.DiskUsed, info.DiskTotal)
	}

	// Timestamp should be recent
	if info.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	t.Logf("Collected system info: %+v", info)
}

func TestCollector_IPAddresses(t *testing.T) {
	collector := NewCollector()

	info, err := collector.Collect()
	if err != nil {
		t.Fatal(err)
	}

	// Should have at least one IP address in most environments
	// (Some test environments might not have any)
	if len(info.IPAddresses) > 0 {
		for _, ip := range info.IPAddresses {
			if ip == "" {
				t.Error("IP address should not be empty string")
			}
			// Basic validation that it looks like an IP
			if len(ip) < 7 { // Minimum: "1.2.3.4"
				t.Errorf("IP address looks invalid: %s", ip)
			}
		}
		t.Logf("Found IP addresses: %v", info.IPAddresses)
	} else {
		t.Log("No IP addresses found (may be normal in test environment)")
	}
}

func BenchmarkCollector_Collect(b *testing.B) {
	collector := NewCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := collector.Collect()
		if err != nil {
			b.Fatal(err)
		}
	}
}
