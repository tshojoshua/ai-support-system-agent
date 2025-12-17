package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once sync.Once
	reg  *prometheus.Registry
)

// Metrics holds all Prometheus metrics for the agent
type Metrics struct {
	// Counter metrics
	HeartbeatTotal         *prometheus.CounterVec
	JobsExecutedTotal      *prometheus.CounterVec
	EnrollmentAttemptsTotal *prometheus.CounterVec
	PolicyViolationsTotal  *prometheus.CounterVec
	UpdateAttemptsTotal    *prometheus.CounterVec
	CertRotationTotal      *prometheus.CounterVec

	// Gauge metrics
	AgentUp                      *prometheus.GaugeVec
	HeartbeatLastSuccess         prometheus.Gauge
	JobExecutionActive           prometheus.Gauge
	HubConnectionStatus          *prometheus.GaugeVec
	PolicyExpirationTimestamp    prometheus.Gauge
	CertExpirationTimestamp      prometheus.Gauge
	SystemCPUUsagePercent        prometheus.Gauge
	SystemMemoryUsedBytes        prometheus.Gauge
	SystemDiskUsedBytes          prometheus.Gauge

	// Histogram metrics
	HeartbeatDuration        prometheus.Histogram
	JobExecutionDuration     *prometheus.HistogramVec
	ArtifactUploadDuration   prometheus.Histogram
	APIRequestDuration       *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(version string) *Metrics {
	once.Do(func() {
		reg = prometheus.NewRegistry()
	})

	m := &Metrics{
		// Counters
		HeartbeatTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "jtnt_agent_heartbeat_total",
				Help: "Total number of heartbeat attempts",
			},
			[]string{"status"},
		),

		JobsExecutedTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "jtnt_agent_jobs_executed_total",
				Help: "Total number of jobs executed",
			},
			[]string{"type", "status"},
		),

		EnrollmentAttemptsTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "jtnt_agent_enrollment_attempts_total",
				Help: "Total number of enrollment attempts",
			},
			[]string{"status"},
		),

		PolicyViolationsTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "jtnt_agent_policy_violations_total",
				Help: "Total number of policy violations",
			},
			[]string{"type"},
		),

		UpdateAttemptsTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "jtnt_agent_update_attempts_total",
				Help: "Total number of update attempts",
			},
			[]string{"status"},
		),

		CertRotationTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "jtnt_agent_cert_rotation_total",
				Help: "Total number of certificate rotation attempts",
			},
			[]string{"status"},
		),

		// Gauges
		AgentUp: promauto.With(reg).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_up",
				Help: "Agent running status",
			},
			[]string{"version"},
		),

		HeartbeatLastSuccess: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_heartbeat_last_success_timestamp",
				Help: "Timestamp of last successful heartbeat",
			},
		),

		JobExecutionActive: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_job_execution_active",
				Help: "Number of jobs currently executing",
			},
		),

		HubConnectionStatus: promauto.With(reg).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_hub_connection_status",
				Help: "Hub connection status (1=connected, 0=disconnected)",
			},
			[]string{"status"},
		),

		PolicyExpirationTimestamp: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_policy_expiration_timestamp",
				Help: "Policy expiration timestamp",
			},
		),

		CertExpirationTimestamp: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_cert_expiration_timestamp",
				Help: "Certificate expiration timestamp",
			},
		),

		SystemCPUUsagePercent: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_system_cpu_usage_percent",
				Help: "System CPU usage percentage",
			},
		),

		SystemMemoryUsedBytes: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_system_memory_used_bytes",
				Help: "System memory used in bytes",
			},
		),

		SystemDiskUsedBytes: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "jtnt_agent_system_disk_used_bytes",
				Help: "System disk used in bytes",
			},
		),

		// Histograms
		HeartbeatDuration: promauto.With(reg).NewHistogram(
			prometheus.HistogramOpts{
				Name:    "jtnt_agent_heartbeat_duration_seconds",
				Help:    "Duration of heartbeat operations",
				Buckets: prometheus.DefBuckets,
			},
		),

		JobExecutionDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "jtnt_agent_job_execution_duration_seconds",
				Help:    "Duration of job execution",
				Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 120, 300, 600},
			},
			[]string{"type"},
		),

		ArtifactUploadDuration: promauto.With(reg).NewHistogram(
			prometheus.HistogramOpts{
				Name:    "jtnt_agent_artifact_upload_duration_seconds",
				Help:    "Duration of artifact uploads",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
			},
		),

		APIRequestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "jtnt_agent_api_request_duration_seconds",
				Help:    "Duration of API requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
	}

	// Set agent up with version
	m.AgentUp.WithLabelValues(version).Set(1)

	return m
}

// GetRegistry returns the Prometheus registry
func GetRegistry() *prometheus.Registry {
	return reg
}

// RecordHeartbeat records a heartbeat attempt
func (m *Metrics) RecordHeartbeat(success bool, duration time.Duration) {
	status := "success"
	if !success {
		status = "error"
	}
	m.HeartbeatTotal.WithLabelValues(status).Inc()
	m.HeartbeatDuration.Observe(duration.Seconds())
	
	if success {
		m.HeartbeatLastSuccess.SetToCurrentTime()
		m.HubConnectionStatus.WithLabelValues("connected").Set(1)
		m.HubConnectionStatus.WithLabelValues("disconnected").Set(0)
	} else {
		m.HubConnectionStatus.WithLabelValues("connected").Set(0)
		m.HubConnectionStatus.WithLabelValues("disconnected").Set(1)
	}
}

// RecordJobExecution records a job execution
func (m *Metrics) RecordJobExecution(jobType string, status string, duration time.Duration) {
	m.JobsExecutedTotal.WithLabelValues(jobType, status).Inc()
	m.JobExecutionDuration.WithLabelValues(jobType).Observe(duration.Seconds())
}

// RecordPolicyViolation records a policy violation
func (m *Metrics) RecordPolicyViolation(violationType string) {
	m.PolicyViolationsTotal.WithLabelValues(violationType).Inc()
}

// RecordEnrollment records an enrollment attempt
func (m *Metrics) RecordEnrollment(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.EnrollmentAttemptsTotal.WithLabelValues(status).Inc()
}

// RecordUpdate records an update attempt
func (m *Metrics) RecordUpdate(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.UpdateAttemptsTotal.WithLabelValues(status).Inc()
}

// RecordCertRotation records a certificate rotation attempt
func (m *Metrics) RecordCertRotation(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	m.CertRotationTotal.WithLabelValues(status).Inc()
}

// SetJobActive sets the active job count
func (m *Metrics) SetJobActive(active bool) {
	if active {
		m.JobExecutionActive.Set(1)
	} else {
		m.JobExecutionActive.Set(0)
	}
}

// UpdateSystemMetrics updates system resource metrics
func (m *Metrics) UpdateSystemMetrics(cpuPercent float64, memoryBytes uint64, diskBytes uint64) {
	m.SystemCPUUsagePercent.Set(cpuPercent)
	m.SystemMemoryUsedBytes.Set(float64(memoryBytes))
	m.SystemDiskUsedBytes.Set(float64(diskBytes))
}

// SetPolicyExpiration sets the policy expiration timestamp
func (m *Metrics) SetPolicyExpiration(expiresAt time.Time) {
	m.PolicyExpirationTimestamp.Set(float64(expiresAt.Unix()))
}

// SetCertExpiration sets the certificate expiration timestamp
func (m *Metrics) SetCertExpiration(expiresAt time.Time) {
	m.CertExpirationTimestamp.Set(float64(expiresAt.Unix()))
}
