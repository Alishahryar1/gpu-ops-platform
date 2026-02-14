// Package config provides configuration management for the GPU Ops Platform.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds the daemon configuration.
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Health HealthConfig `mapstructure:"health"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	Policy PolicyConfig `mapstructure:"policy"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// HealthConfig holds health monitoring configuration.
type HealthConfig struct {
	CheckInterval    time.Duration `mapstructure:"check_interval"`
	RegistryPath     string        `mapstructure:"registry_path"`
	FailedThreshold  int           `mapstructure:"failed_threshold"`
	AutoUnregister   bool          `mapstructure:"auto_unregister"`
}

// MetricsConfig holds Prometheus metrics configuration.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

// PolicyConfig holds policy engine configuration.
type PolicyConfig struct {
	Engine         string        `mapstructure:"engine"` // "starlark", "none"
	PolicyDir      string        `mapstructure:"policy_dir"`
	ReloadInterval time.Duration `mapstructure:"reload_interval"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // "json", "text"
	Output string `mapstructure:"output"` // "stdout", file path
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Health: HealthConfig{
			CheckInterval:    10 * time.Second,
			RegistryPath:     "/var/lib/gputl/registry.json",
			FailedThreshold:  3,
			AutoUnregister:   false,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
		Policy: PolicyConfig{
			Engine:         "starlark",
			PolicyDir:      "/etc/gputl/policies",
			ReloadInterval: 60 * time.Second,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

// Load loads configuration from file and environment variables.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	config := Default()
	setDefaults(v)

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	} else {
		// Try default locations
		v.SetConfigName("gputl")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/gputl/")
		v.AddConfigPath("$HOME/.config/gputl/")
		v.AddConfigPath(".")

		// Optionally read config file (don't fail if not found)
		v.ReadInConfig()
	}

	// Bind environment variables
	v.SetEnvPrefix("GPUTL")
	v.AutomaticEnv()

	// Unmarshal to config
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) {
	config := Default()

	v.SetDefault("server.host", config.Server.Host)
	v.SetDefault("server.port", config.Server.Port)
	v.SetDefault("server.read_timeout", config.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", config.Server.WriteTimeout)

	v.SetDefault("health.check_interval", config.Health.CheckInterval)
	v.SetDefault("health.registry_path", config.Health.RegistryPath)
	v.SetDefault("health.failed_threshold", config.Health.FailedThreshold)
	v.SetDefault("health.auto_unregister", config.Health.AutoUnregister)

	v.SetDefault("metrics.enabled", config.Metrics.Enabled)
	v.SetDefault("metrics.port", config.Metrics.Port)
	v.SetDefault("metrics.path", config.Metrics.Path)

	v.SetDefault("policy.engine", config.Policy.Engine)
	v.SetDefault("policy.policy_dir", config.Policy.PolicyDir)
	v.SetDefault("policy.reload_interval", config.Policy.ReloadInterval)

	v.SetDefault("logging.level", config.Logging.Level)
	v.SetDefault("logging.format", config.Logging.Format)
	v.SetDefault("logging.output", config.Logging.Output)
}

// Write creates a sample configuration file.
func Write(path string) error {
	config := Default()

	content := `# GPU Ops Platform Configuration

server:
  host: ` + config.Server.Host + `
  port: ` + toString(config.Server.Port) + `
  read_timeout: ` + config.Server.ReadTimeout.String() + `
  write_timeout: ` + config.Server.WriteTimeout.String() + `

health:
  check_interval: ` + config.Health.CheckInterval.String() + `
  registry_path: ` + config.Health.RegistryPath + `
  failed_threshold: ` + toString(config.Health.FailedThreshold) + `
  auto_unregister: ` + toBoolString(config.Health.AutoUnregister) + `

metrics:
  enabled: ` + toBoolString(config.Metrics.Enabled) + `
  port: ` + toString(config.Metrics.Port) + `
  path: ` + config.Metrics.Path + `

policy:
  engine: ` + config.Policy.Engine + `
  policy_dir: ` + config.Policy.PolicyDir + `
  reload_interval: ` + config.Policy.ReloadInterval.String() + `

logging:
  level: ` + config.Logging.Level + `
  format: ` + config.Logging.Format + `
  output: ` + config.Logging.Output + `
`
	return os.WriteFile(path, []byte(content), 0644)
}

func toString(i int) string {
	return fmt.Sprintf("%d", i)
}

func toBoolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
