// gputld is the GPU Ops Platform daemon.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alish/gpu-ops-platform/pkg/health"
	"github.com/alish/gpu-ops-platform/pkg/metrics"
	"github.com/alish/gpu-ops-platform/pkg/registration"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultPort           = 8080
	metricsPort           = 9090
	defaultHealthInterval = 10 * time.Second
)

type GPUManager struct{}

// GetGPU returns a mock GPU for development.
func (m *GPUManager) GetGPU(id int) (*health.GPU, error) {
	return &health.GPU{
		ID:            id,
		Name:          "NVIDIA GeForce RTX 5070 Ti",
		TemperatureC:  35.0,
		PowerUsageW:   10.0,
		PowerLimitW:   250.0,
		UsedMemoryGB:  0.0,
		TotalMemoryGB: 24.0,
		Online:        true,
		Registered:    true,
	}, nil
}

// GetAllGPUs returns all GPUs.
func (m *GPUManager) GetAllGPUs() []health.GPU {
	return []health.GPU{
		{
			ID:            0,
			Name:          "NVIDIA GeForce RTX 5070 Ti",
			TemperatureC:  35.0,
			PowerUsageW:   10.0,
			PowerLimitW:   250.0,
			UsedMemoryGB:  0.0,
			TotalMemoryGB: 24.0,
			Online:        true,
			Registered:    true,
		},
	}
}

type Daemon struct {
	registry       *registration.Registry
	healthMonitor  *health.HealthMonitor
	gpuManager     *GPUManager
	httpServer     *http.Server
	metricsServer  *http.Server
	healthInterval time.Duration
}

func NewDaemon() *Daemon {
	gpuManager := &GPUManager{}

	return &Daemon{
		registry:       registration.NewRegistry(),
		healthMonitor:  health.NewHealthMonitor(gpuManager),
		gpuManager:     gpuManager,
		healthInterval: defaultHealthInterval,
	}
}

func (d *Daemon) Initialize() error {
	log.Println("Initializing GPU Ops Platform daemon...")

	// Load existing registry if exists
	if err := d.registry.LoadFromFile("/var/lib/gputl/registry.json"); err != nil {
		log.Printf("Warning: Could not load registry: %v", err)
	}

	// Register default health checks
	d.healthMonitor.RegisterCheck(&health.CheckConfig{
		Name:              "temperature_check",
		CheckType:         "temperature",
		IntervalSec:       10,
		WarningThreshold:  75.0,
		CriticalThreshold: 85.0,
		Enabled:           true,
	})

	d.healthMonitor.RegisterCheck(&health.CheckConfig{
		Name:              "power_check",
		CheckType:         "power",
		IntervalSec:       60,
		WarningThreshold:  80.0,
		CriticalThreshold: 95.0,
		Enabled:           true,
	})

	d.healthMonitor.RegisterCheck(&health.CheckConfig{
		Name:              "memory_check",
		CheckType:         "memory",
		IntervalSec:       30,
		WarningThreshold:  80.0,
		CriticalThreshold: 95.0,
		Enabled:           true,
	})

	return nil
}

func (d *Daemon) StartHTTPServer(port int) {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/health", d.handleHealth)
	mux.HandleFunc("/api/v1/gpus", d.handleListGPUs)
	mux.HandleFunc("/api/v1/gpus/", d.handleGetGPU)
	mux.HandleFunc("/api/v1/register", d.handleRegister)
	mux.HandleFunc("/api/v1/unregister/", d.handleUnregister)
	mux.HandleFunc("/api/v1/healthchecks", d.handleHealthChecks)

	d.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("HTTP server listening on port %d", port)

	go func() {
		if err := d.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
}

func (d *Daemon) StartMetricsServer(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	d.metricsServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("Metrics server listening on port %d", port)

	go func() {
		if err := d.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Metrics server error: %v", err)
		}
	}()
}

func (d *Daemon) RunHealthChecks(ctx context.Context) {
	ticker := time.NewTicker(d.healthInterval)
	defer ticker.Stop()

	log.Println("Starting health check loop...")

	for {
		select {
		case <-ticker.C:
			gpus := d.gpuManager.GetAllGPUs()
			for _, gpu := range gpus {
				if gpu.Online && gpu.Registered {
					checks := d.healthMonitor.ListChecks()
					for _, check := range checks {
						if check.Enabled {
							result, err := d.healthMonitor.RunCheck(check.Name, gpu.ID)
							if err != nil {
								log.Printf("Health check error for GPU %d: %v", gpu.ID, err)
								continue
							}

							// Update Prometheus metrics
							metrics.UpdateHealthCheckStatus(gpu.ID, check.Name, string(result.Status))

							if result.Status == health.StatusUnhealthy {
								log.Printf("ALERT: GPU %d - %s", gpu.ID, result.Message)
							}
						}
					}
				}
			}
		case <-ctx.Done():
			log.Println("Health check loop stopped")
			return
		}
	}
}

func (d *Daemon) Shutdown(ctx context.Context) error {
	log.Println("Shutting down daemon...")

	// Graceful shutdown of servers
	if d.httpServer != nil {
		if err := d.httpServer.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}

	if d.metricsServer != nil {
		if err := d.metricsServer.Shutdown(ctx); err != nil {
			log.Printf("Metrics server shutdown error: %v", err)
		}
	}

	// Save registry state
	if err := d.registry.SaveToFile("/var/lib/gputl/registry.json"); err != nil {
		log.Printf("Warning: Could not save registry: %v", err)
	}

	// Stop health monitor
	d.healthMonitor.Stop()

	log.Println("Daemon shutdown complete")
	return nil
}

// HTTP Handlers

func (d *Daemon) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func (d *Daemon) handleListGPUs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gpus := d.gpuManager.GetAllGPUs()
	for _, gpu := range gpus {
		fmt.Fprintf(w, "GPU%d: %s (Online: %v, Registered: %v)\n", gpu.ID, gpu.Name, gpu.Online, gpu.Registered)
	}
}

func (d *Daemon) handleGetGPU(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract GPU ID from path (simplified)
	fmt.Fprintf(w, "GPU details endpoint\n")
}

func (d *Daemon) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Parse request body and register GPU
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GPU registered\n"))
}

func (d *Daemon) handleUnregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Extract GPU ID and unregister
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("GPU unregistered\n"))
}

func (d *Daemon) handleHealthChecks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checks := d.healthMonitor.ListChecks()
	for _, check := range checks {
		fmt.Fprintf(w, "%s: enabled=%v, interval=%ds\n", check.Name, check.Enabled, check.IntervalSec)
	}
}

func main() {
	daemon := NewDaemon()

	if err := daemon.Initialize(); err != nil {
		log.Fatalf("Failed to initialize daemon: %v", err)
	}

	// Start services
	daemon.StartHTTPServer(defaultPort)
	daemon.StartMetricsServer(metricsPort)

	// Start health checks in background
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go daemon.RunHealthChecks(ctx)

	// Wait for shutdown signal
	<-ctx.Done()
	daemon.Shutdown(context.Background())
}
