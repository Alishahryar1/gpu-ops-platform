// Package health handles GPU health checks and status monitoring.
package health

import (
	"fmt"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// CheckConfig defines configuration for a health check.
type CheckConfig struct {
	Name              string
	CheckType         string // "temperature", "power", "memory", "pci"
	IntervalSec       int
	WarningThreshold  float64
	CriticalThreshold float64
	Enabled           bool
}

// CheckResult represents the result of a single health check.
type CheckResult struct {
	CheckName string
	GPUID     int
	Status    Status
	Value     float64
	Threshold float64
	Message   string
	Timestamp time.Time
}

// HealthMonitor monitors GPU health across all registered checks.
type HealthMonitor struct {
	mu         sync.RWMutex
	checks     map[string]*CheckConfig
	results    map[string]*CheckResult
	gpuManager GPUInterface
	stopped    chan struct{}
}

// GPUInterface defines the interface for GPU operations.
type GPUInterface interface {
	GetAllGPUs() []GPU
	GetGPU(id int) (*GPU, error)
}

// GPU represents a minimal GPU interface for health checks.
type GPU struct {
	ID            int
	Name          string
	TemperatureC  float64
	PowerUsageW   float64
	PowerLimitW   float64
	UsedMemoryGB  float64
	TotalMemoryGB float64
	Online        bool
	Registered    bool
}

// NewHealthMonitor creates a new health monitor.
func NewHealthMonitor(gpuManager GPUInterface) *HealthMonitor {
	return &HealthMonitor{
		checks:     make(map[string]*CheckConfig),
		results:    make(map[string]*CheckResult),
		gpuManager: gpuManager,
		stopped:    make(chan struct{}),
	}
}

// RegisterCheck registers a new health check.
func (hm *HealthMonitor) RegisterCheck(config *CheckConfig) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.checks[config.Name] = config
}

// GetCheck retrieves a check configuration by name.
func (hm *HealthMonitor) GetCheck(name string) (*CheckConfig, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	check, ok := hm.checks[name]
	if !ok {
		return nil, fmt.Errorf("check %s not found", name)
	}
	return check, nil
}

// ListChecks returns all registered checks.
func (hm *HealthMonitor) ListChecks() []*CheckConfig {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	checks := make([]*CheckConfig, 0, len(hm.checks))
	for _, check := range hm.checks {
		checks = append(checks, check)
	}
	return checks
}

// RunCheck executes a single health check for a specific GPU.
func (hm *HealthMonitor) RunCheck(checkName string, gpuID int) (*CheckResult, error) {
	hm.mu.RLock()
	check, ok := hm.checks[checkName]
	hm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("check %s not registered", checkName)
	}

	if !check.Enabled {
		return nil, fmt.Errorf("check %s is disabled", checkName)
	}

	// Get GPU data
	gpu, err := hm.gpuManager.GetGPU(gpuID)
	if err != nil {
		return &CheckResult{
			CheckName: checkName,
			GPUID:     gpuID,
			Status:    StatusUnhealthy,
			Message:   fmt.Sprintf("GPU error: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	result := &CheckResult{
		CheckName: checkName,
		GPUID:     gpuID,
		Timestamp: time.Now(),
	}

	// Execute check based on type
	switch check.CheckType {
	case "temperature":
		result = hm.checkTemperature(check, gpuID, gpu.TemperatureC)
	case "power":
		result = hm.checkPower(check, gpuID, gpu.PowerUsageW, gpu.PowerLimitW)
	case "memory":
		result = hm.checkMemory(check, gpuID, gpu.UsedMemoryGB, gpu.TotalMemoryGB)
	default:
		result.Status = StatusUnknown
		result.Message = fmt.Sprintf("Unknown check type: %s", check.CheckType)
	}

	hm.mu.Lock()
	hm.results[fmt.Sprintf("%s:%d", checkName, gpuID)] = result
	hm.mu.Unlock()

	return result, nil
}

// checkTemperature evaluates temperature health.
func (hm *HealthMonitor) checkTemperature(config *CheckConfig, gpuID int, tempC float64) *CheckResult {
	result := &CheckResult{
		CheckName: config.Name,
		GPUID:     gpuID,
		Value:     tempC,
		Timestamp: time.Now(),
	}

	result.Threshold = config.CriticalThreshold

	switch {
	case tempC >= config.CriticalThreshold:
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Critical temperature: %.1f°C >= %.1f°C", tempC, config.CriticalThreshold)
	case tempC >= config.WarningThreshold:
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("High temperature warning: %.1f°C >= %.1f°C", tempC, config.WarningThreshold)
	default:
		result.Status = StatusHealthy
		result.Message = fmt.Sprintf("Temperature normal: %.1f°C", tempC)
	}

	return result
}

// checkPower evaluates power usage health.
func (hm *HealthMonitor) checkPower(config *CheckConfig, gpuID int, powerW, limitW float64) *CheckResult {
	result := &CheckResult{
		CheckName: config.Name,
		GPUID:     gpuID,
		Value:     powerW,
		Timestamp: time.Now(),
	}

	percentOfLimit := (powerW / limitW) * 100
	result.Threshold = config.WarningThreshold

	switch {
	case percentOfLimit >= config.CriticalThreshold:
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Critical power usage: %.1fW (%.1f%% of %.1fW limit)", powerW, percentOfLimit, limitW)
	case percentOfLimit >= config.WarningThreshold:
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("High power usage: %.1fW (%.1f%% of %.1fW limit)", powerW, percentOfLimit, limitW)
	default:
		result.Status = StatusHealthy
		result.Message = fmt.Sprintf("Power usage normal: %.1fW (%.1f%% of %.1fW limit)", powerW, percentOfLimit, limitW)
	}

	return result
}

// checkMemory evaluates memory health.
func (hm *HealthMonitor) checkMemory(config *CheckConfig, gpuID int, usedGB, totalGB float64) *CheckResult {
	result := &CheckResult{
		CheckName: config.Name,
		GPUID:     gpuID,
		Value:     usedGB,
		Timestamp: time.Now(),
	}

	percentUsed := (usedGB / totalGB) * 100
	result.Threshold = 100 - config.WarningThreshold

	// Invert logic for memory (we warn when high)
	availablePercent := 100 - percentUsed

	switch {
	case availablePercent <= (100 - config.CriticalThreshold):
		result.Status = StatusUnhealthy
		result.Message = fmt.Sprintf("Critical memory usage: %.1f/%.1fGB (%.1f%% used)", usedGB, totalGB, percentUsed)
	case availablePercent <= (100 - config.WarningThreshold):
		result.Status = StatusDegraded
		result.Message = fmt.Sprintf("High memory usage: %.1f/%.1fGB (%.1f%% used)", usedGB, totalGB, percentUsed)
	default:
		result.Status = StatusHealthy
		result.Message = fmt.Sprintf("Memory usage normal: %.1f/%.1fGB (%.1f%% used)", usedGB, totalGB, percentUsed)
	}

	return result
}

// GetLatestResult returns the latest result for a check/GPU combination.
func (hm *HealthMonitor) GetLatestResult(checkName string, gpuID int) (*CheckResult, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	result, ok := hm.results[fmt.Sprintf("%s:%d", checkName, gpuID)]
	if !ok {
		return nil, fmt.Errorf("no result found for check %s on GPU %d", checkName, gpuID)
	}
	return result, nil
}

// GetAllResults returns all health check results.
func (hm *HealthMonitor) GetAllResults() map[string]*CheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	results := make(map[string]*CheckResult, len(hm.results))
	for k, v := range hm.results {
		results[k] = v
	}
	return results
}

// Start starts the health monitoring loop.
func (hm *HealthMonitor) Start() {
	for {
		select {
		case <-hm.stopped:
			return
		default:
			hm.mu.RLock()
			checks := hm.ListChecks()
			hm.mu.RUnlock()

			for _, check := range checks {
				if !check.Enabled {
					continue
				}

			}
			time.Sleep(5 * time.Second)
		}
	}
}

// Stop stops the health monitor.
func (hm *HealthMonitor) Stop() {
	close(hm.stopped)
}
