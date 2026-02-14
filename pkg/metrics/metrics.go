// Package metrics provides Prometheus metric collection.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// GPUMetrics holds all GPU-related Prometheus metrics
	GPUUtilization = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_utilization_percent",
			Help: "Current GPU utilization percentage",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPUTemperatureC = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_temperature_celsius",
			Help: "Current GPU temperature in Celsius",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPUAverageTemperatureC = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gpu_average_temperature_celsius",
			Help: "Average GPU temperature across all active GPUs",
		},
	)

	GPUPowerUsageW = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_power_usage_watts",
			Help: "Current GPU power usage in watts",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPUPowerLimitW = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_power_limit_watts",
			Help: "GPU power limit in watts",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPUMemoryUsedGB = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_memory_used_gb",
			Help: "GPU memory used in GB",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPUMemoryTotalGB = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_memory_total_gb",
			Help: "Total GPU memory in GB",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPUOnline = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_online",
			Help: "GPU online status (1 = online, 0 = offline)",
		},
		[]string{"gpu_id", "gpu_name"},
	)

	GPURegistered = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gpu_registered",
			Help: "GPU registration status (1 = registered, 0 = unregistered)",
		},
		[]string{"gpu_id", "gpu_name", "pool"},
	)

	// Health check metrics
	HealthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "health_check_status",
			Help: "Health check status (1 = healthy, 0.5 = degraded, 0 = unhealthy)",
		},
		[]string{"gpu_id", "check_name"},
	)

	HealthCheckLatencySeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "health_check_latency_seconds",
			Help:    "Health check execution time",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"gpu_id", "check_name"},
	)

	// API metrics
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	APILatencySeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_latency_seconds",
			Help:    "API request latency",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

// GPUInfo contains GPU metrics for Prometheus.
type GPUInfo struct {
	ID                  int
	Name                string
	UUID                string
	UtilizationPercent  float64
	TemperatureC        float64
	PowerUsageW         float64
	PowerLimitW         float64
	MemoryUsedGB        float64
	MemoryTotalGB       float64
	Online              bool
	Registered          bool
	Pool                string
}

// UpdateGPUMetrics updates Prometheus metrics for a GPU.
func UpdateGPUMetrics(gpu GPUInfo) {
	labels := prometheus.Labels{
		"gpu_id":   string(rune(gpu.ID + '0')),
		"gpu_name": gpu.Name,
	}

	GPUUtilization.With(labels).Set(gpu.UtilizationPercent)
	GPUTemperatureC.With(labels).Set(gpu.TemperatureC)
	GPUPowerUsageW.With(labels).Set(gpu.PowerUsageW)
	GPUPowerLimitW.With(labels).Set(gpu.PowerLimitW)
	GPUMemoryUsedGB.With(labels).Set(gpu.MemoryUsedGB)
	GPUMemoryTotalGB.With(labels).Set(gpu.MemoryTotalGB)

	if gpu.Online {
		GPUOnline.With(labels).Set(1)
	} else {
		GPUOnline.With(labels).Set(0)
	}

	if gpu.Registered {
		GPURegistered.With(prometheus.Labels{
			"gpu_id":   string(rune(gpu.ID + '0')),
			"gpu_name": gpu.Name,
			"pool":     gpu.Pool,
		}).Set(1)
	}
}

// UpdateHealthCheckStatus updates the health check status metric.
func UpdateHealthCheckStatus(gpuID int, checkName string, status string) {
	labels := prometheus.Labels{
		"gpu_id":     string(rune(gpuID + '0')),
		"check_name": checkName,
	}

	var value float64
	switch status {
	case "healthy":
		value = 1.0
	case "degraded":
		value = 0.5
	case "unhealthy":
		value = 0.0
	default:
		value = -1.0 // unknown
	}

	HealthCheckStatus.With(labels).Set(value)
}

// RecordAPIMetrics records API request metrics.
func RecordAPIMetrics(method, endpoint, status string, latency float64) {
	APIRequestsTotal.With(prometheus.Labels{
		"method":   method,
		"endpoint": endpoint,
		"status":   status,
	}).Inc()

	APILatencySeconds.With(prometheus.Labels{
		"method":   method,
		"endpoint": endpoint,
	}).Observe(latency)
}
