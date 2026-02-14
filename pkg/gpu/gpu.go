// Package gpu provides GPU discovery and monitoring functionality.
package gpu

import (
	"fmt"
	"sync"
	"time"
)

// GPU represents a single GPU device.
type GPU struct {
	ID               int
	Name             string
	UUID             string
	TotalMemoryGB    float64
	UsedMemoryGB     float64
	UtilizationPercent float64
	TemperatureC     float64
	PowerUsageW      float64
	PowerLimitW      float64
	ClockMHz         float64
	Online           bool
	Registered       bool
	RegisteredAt     *time.Time
	LastUpdated      time.Time
}

// Info returns GPU information as formatted string.
func (g *GPU) Info() string {
	return fmt.Sprintf(
		"GPU%d [%s] - %s | Mem: %.1f/%.1fGB | Temp: %.1f°C | Util: %.1f%%",
		g.ID, g.UUID, g.Name, g.UsedMemoryGB, g.TotalMemoryGB, g.TemperatureC, g.UtilizationPercent,
	)
}

// Manager handles GPU discovery, monitoring, and state management.
type Manager struct {
	mu      sync.RWMutex
	gpus    map[int]*GPU
	stopped chan struct{}
}

// NewManager creates a new GPU manager.
func NewManager() *Manager {
	return &Manager{
		gpus:    make(map[int]*GPU),
		stopped: make(chan struct{}),
	}
}

// Discover discovers all available GPUs on the system.
func (m *Manager) Discover() ([]*GPU, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// NOTE: Replace this with actual NVML calls when integrated
	// For now, return mock data for development
	gpus := []*GPU{
		{
			ID:            0,
			Name:          "NVIDIA GeForce RTX 5070 Ti",
			UUID:          "GPU-12345678-90ab-cdef-1234-567890abcdef",
			TotalMemoryGB: 24.0,
			UsedMemoryGB:  0.0,
			Online:        true,
			LastUpdated:   time.Now(),
		},
	}

	for _, gpu := range gpus {
		m.gpus[gpu.ID] = gpu
	}

	return gpus, nil
}

// GetGPU retrieves a GPU by ID.
func (m *Manager) GetGPU(id int) (*GPU, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gpu, ok := m.gpus[id]
	if !ok {
		return nil, fmt.Errorf("GPU %d not found", id)
	}

	return &*gpu, nil
}

// GetAllGPUs returns all GPUs.
func (m *Manager) GetAllGPUs() []*GPU {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gpus := make([]*GPU, 0, len(m.gpus))
	for _, gpu := range m.gpus {
		gpus = append(gpus, &*gpu)
	}
	return gpus
}

// UpdateMetrics updates GPU metrics (called periodically).
func (m *Manager) UpdateMetrics() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// NOTE: Replace with actual NVML queries
	for _, gpu := range m.gpus {
		// Mock metrics update
		gpu.TemperatureC = 35.0
		gpu.UtilizationPercent = 0.0
		gpu.PowerUsageW = 10.0
		gpu.LastUpdated = time.Now()
	}

	return nil
}

// Register marks a GPU as registered.
func (m *Manager) Register(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gpu, ok := m.gpus[id]
	if !ok {
		return fmt.Errorf("GPU %d not found", id)
	}

	gpu.Registered = true
	now := time.Now()
	gpu.RegisteredAt = &now
	return nil
}

// Unregister marks a GPU as unregistered.
func (m *Manager) Unregister(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gpu, ok := m.gpus[id]
	if !ok {
		return fmt.Errorf("GPU %d not found", id)
	}

	gpu.Registered = false
	gpu.RegisteredAt = nil
	return nil
}

// StartMonitoring starts periodic GPU monitoring.
func (m *Manager) StartMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.UpdateMetrics(); err != nil {
				fmt.Printf("GPU metrics update error: %v\n", err)
			}
		case <-m.stopped:
			return
		}
	}
}

// Stop stops the GPU manager.
func (m *Manager) Stop() {
	close(m.stopped)
}
