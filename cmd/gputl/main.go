// gputl is the GPU Ops Platform CLI tool.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDaemonURL = "http://localhost:8080"
	defaultTimeout   = 30 * time.Second
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultDaemonURL
	}

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) DoRequest(method, path string, body io.Reader) ([]byte, error) {
	url := c.BaseURL + path
	resp, err := c.HTTPClient.DoRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return data, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// Commands

func cmdStatus(client *Client, gpuID string) error {
	if gpuID != "" {
		id, err := strconv.Atoi(gpuID)
		if err != nil {
			return fmt.Errorf("invalid GPU ID: %v", err)
		}
		return getGPUStatus(client, id)
	}
	return listGPUs(client)
}

func listGPUs(client *Client) error {
	data, err := client.DoRequest("GET", "/api/v1/gpus", nil)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func getGPUStatus(client *Client, gpuID int) error {
	data, err := client.DoRequest("GET", fmt.Sprintf("/api/v1/gpus/%d", gpuID), nil)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func cmdRegister(client *Client, gpuID, name, pool string, tags []string) error {
	id, err := strconv.Atoi(gpuID)
	if err != nil {
		return fmt.Errorf("invalid GPU ID: %v", err)
	}

	req := map[string]interface{}{
		"gpu_id": id,
		"name":   name,
		"pool":   pool,
		"tags":   tags,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	data, err := client.DoRequest("POST", "/api/v1/register", strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func cmdUnregister(client *Client, gpuID string) error {
	id, err := strconv.Atoi(gpuID)
	if err != nil {
		return fmt.Errorf("invalid GPU ID: %v", err)
	}

	data, err := client.DoRequest("DELETE", fmt.Sprintf("/api/v1/unregister/%d", id), nil)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func cmdHealthChecks(client *Client) error {
	data, err := client.DoRequest("GET", "/api/v1/healthchecks", nil)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func cmdHealth(client *Client) error {
	resp, err := client.HTTPClient.Get(client.BaseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Daemon is healthy")
		return nil
	}

	return fmt.Errorf("daemon is unhealthy: status %d", resp.StatusCode)
}

func printUsage() {
	fmt.Println("GPU Ops Platform CLI (gputl)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gputl <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  status [gpu-id]         Show GPU status (all GPUs or specific GPU)")
	fmt.Println("  register <gpu-id> [options]  Register a GPU")
	fmt.Println("  unregister <gpu-id>     Unregister a GPU")
	fmt.Println("  health-checks          List health check configurations")
	fmt.Println("  health                  Check daemon health")
	fmt.Println()
	fmt.Println("Register Options:")
	fmt.Println("  --name <name>          GPU name")
	fmt.Println("  --pool <pool>          GPU pool (default: default)")
	fmt.Println("  --tags <tag1,tag2>     Comma-separated tags")
	fmt.Println()
	fmt.Println("Global Options:")
	fmt.Println("  --daemon-url <url>     Daemon URL (default: http://localhost:8080)")
	fmt.Println("  --help                 Show help")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Parse global options
	daemonURL := defaultDaemonURL
	args := os.Args[2:]
	var filteredArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--daemon-url=") {
			daemonURL = strings.TrimPrefix(arg, "--daemon-url=")
		} else if arg == "--daemon-url" && i+1 < len(args) {
			daemonURL = args[i+1]
			i++
		} else if arg == "--help" {
			printUsage()
			os.Exit(0)
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	client := NewClient(daemonURL)

	// Execute command
	switch command {
	case "status":
		var gpuID string
		if len(filteredArgs) > 0 {
			gpuID = filteredArgs[0]
		}
		if err := cmdStatus(client, gpuID); err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "register":
		if len(filteredArgs) < 1 {
			log.Fatal("Error: GPU ID required")
		}

		name := ""
		pool := "default"
		tags := []string{}

		for i := 1; i < len(filteredArgs); i += 2 {
			if i+1 >= len(filteredArgs) {
				break
			}
			arg := filteredArgs[i]
			value := filteredArgs[i+1]

			switch arg {
			case "--name":
				name = value
			case "--pool":
				pool = value
			case "--tags":
				tags = strings.Split(value, ",")
			}
		}

		if err := cmdRegister(client, filteredArgs[0], name, pool, tags); err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "unregister":
		if len(filteredArgs) < 1 {
			log.Fatal("Error: GPU ID required")
		}
		if err := cmdUnregister(client, filteredArgs[0]); err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "health-checks":
		if err := cmdHealthChecks(client); err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "health":
		if err := cmdHealth(client); err != nil {
			log.Fatalf("Error: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}
