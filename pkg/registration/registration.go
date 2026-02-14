// Package registration handles GPU registration and lifecycle management.
package registration

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Registry stores GPU registration records.
type Registry struct {
	mu     sync.RWMutex
	records map[string]*GPURecord
}

// GPURecord represents a registered GPU.
type GPURecord struct {
	GPUID          int       `json:"gpu_id"`
	UUID           string    `json:"uuid"`
	Name           string    `json:"name"`
	Pool           string    `json:"pool"`
	Tags           []string  `json:"tags"`
	Metadata       map[string]string `json:"metadata"`
	RegisteredAt   time.Time `json:"registered_at"`
	RegisteredBy   string    `json:"registered_by"`
	LastHealthCheck time.Time `json:"last_health_check"`
	Status         string    `json:"status"` // "active", "inactive", "degraded"
}

// NewRegistry creates a new GPU registry.
func NewRegistry() *Registry {
	return &Registry{
		records: make(map[string]*GPURecord),
	}
}

// Register registers a GPU in the registry.
func (r *Registry) Register(gpuID int, uuid, name, pool string, tags []string, metadata map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d", gpuID)

	if _, exists := r.records[key]; exists {
		return fmt.Errorf("GPU %d is already registered", gpuID)
	}

	r.records[key] = &GPURecord{
		GPUID:        gpuID,
		UUID:         uuid,
		Name:         name,
		Pool:         pool,
		Tags:         tags,
		Metadata:     metadata,
		RegisteredAt: time.Now(),
		RegisteredBy: "gputl",
		Status:       "active",
	}

	return nil
}

// Unregister removes a GPU from the registry.
func (r *Registry) Unregister(gpuID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d", gpuID)

	if _, exists := r.records[key]; !exists {
		return fmt.Errorf("GPU %d is not registered", gpuID)
	}

	delete(r.records, key)
	return nil
}

// Get retrieves a GPU record by ID.
func (r *Registry) Get(gpuID int) (*GPURecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%d", gpuID)
	record, ok := r.records[key]
	if !ok {
		return nil, fmt.Errorf("GPU %d not found in registry", gpuID)
	}

	return record, nil
}

// List returns all registered GPUs.
func (r *Registry) List() []*GPURecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records := make([]*GPURecord, 0, len(r.records))
	for _, record := range r.records {
		records = append(records, record)
	}
	return records
}

// FindByPool returns all GPUs in a specific pool.
func (r *Registry) FindByPool(pool string) []*GPURecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var records []*GPURecord
	for _, record := range r.records {
		if record.Pool == pool {
			records = append(records, record)
		}
	}
	return records
}

// FindByTag returns all GPUs with a specific tag.
func (r *Registry) FindByTag(tag string) []*GPURecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var records []*GPURecord
	for _, record := range r.records {
		for _, t := range record.Tags {
			if t == tag {
				records = append(records, record)
				break
			}
		}
	}
	return records
}

// UpdateStatus updates the status of a GPU.
func (r *Registry) UpdateStatus(gpuID int, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d", gpuID)
	record, ok := r.records[key]
	if !ok {
		return fmt.Errorf("GPU %d not found in registry", gpuID)
	}

	record.Status = status
	record.LastHealthCheck = time.Now()
	return nil
}

// UpdateLastHealthCheck updates the last health check timestamp.
func (r *Registry) UpdateLastHealthCheck(gpuID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("%d", gpuID)
	record, ok := r.records[key]
	if !ok {
		return fmt.Errorf("GPU %d not found in registry", gpuID)
	}

	record.LastHealthCheck = time.Now()
	return nil
}

// SaveToFile saves the registry to a JSON file.
func (r *Registry) SaveToFile(path string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, err := json.MarshalIndent(r.records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// LoadFromFile loads the registry from a JSON file.
func (r *Registry) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Not an error if file doesn't exist yet
		}
		return fmt.Errorf("failed to read registry file: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return json.Unmarshal(data, &r.records)
}
