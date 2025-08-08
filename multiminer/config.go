package multiminer

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the multiminer system
type Config struct {
	Server     ServerConfig     `json:"server"`
	Manager    ManagerConfig    `json:"manager"`
	Pool       PoolConfig       `json:"pool"`
	Logging    LoggingConfig    `json:"logging"`
	Security   SecurityConfig   `json:"security"`
	Validation ValidationConfig `json:"validation"`
}

// ServerConfig configures the HTTP server
type ServerConfig struct {
	ListenAddress string        `json:"listen_address"`
	ReadTimeout   time.Duration `json:"read_timeout"`
	WriteTimeout  time.Duration `json:"write_timeout"`
	IdleTimeout   time.Duration `json:"idle_timeout"`
}

// ManagerConfig configures the device manager
type ManagerConfig struct {
	ProbeTimeout    time.Duration `json:"probe_timeout"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	AutoCleanup     bool          `json:"auto_cleanup"`
}

// PoolConfig configures connection pooling
type PoolConfig struct {
	MaxIdleConnections int           `json:"max_idle_connections"`
	MaxOpenConnections int           `json:"max_open_connections"`
	ConnectionTTL      time.Duration `json:"connection_ttl"`
}

// LoggingConfig configures logging
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	OutputFile string `json:"output_file"`
}

// SecurityConfig configures security settings
type SecurityConfig struct {
	EnableValidation bool     `json:"enable_validation"`
	AllowedPorts     []int    `json:"allowed_ports"`
	AllowedCommands  []string `json:"allowed_commands"`
	RateLimitRPS     int      `json:"rate_limit_rps"`
}

// ValidationConfig configures input validation
type ValidationConfig struct {
	MaxParameterLength int      `json:"max_parameter_length"`
	DangerousChars     []string `json:"dangerous_chars"`
	AllowLocalhost     bool     `json:"allow_localhost"`
	AllowPrivateIPs    bool     `json:"allow_private_ips"`
}

// ManagerOptions holds legacy options for backward compatibility
type ManagerOptions struct {
	ProbeTimeout time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddress: ":8080",
			ReadTimeout:   30 * time.Second,
			WriteTimeout:  30 * time.Second,
			IdleTimeout:   60 * time.Second,
		},
		Manager: ManagerConfig{
			ProbeTimeout:    1200 * time.Millisecond,
			CleanupInterval: 5 * time.Minute,
			AutoCleanup:     true,
		},
		Pool: PoolConfig{
			MaxIdleConnections: 5,
			MaxOpenConnections: 10,
			ConnectionTTL:      5 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			OutputFile: "",
		},
		Security: SecurityConfig{
			EnableValidation: true,
			AllowedPorts:     []int{4028, 8080, 80, 443, 8000, 9090, 3000, 4029},
			AllowedCommands: []string{
				"version", "summary", "devs", "pools", "stats",
				"addpool", "enablepool", "disablepool", "removepool",
				"switchpool", "restart", "quit", "config", "lcd",
				"fans", "temps",
			},
			RateLimitRPS: 100,
		},
		Validation: ValidationConfig{
			MaxParameterLength: 1000,
			DangerousChars:     []string{";", "&", "|", "`", "$", "(", ")", "<", ">"},
			AllowLocalhost:     true,
			AllowPrivateIPs:    true,
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(filename string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return default config
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// LoadConfigWithEnv loads config from file and overrides with environment variables
func LoadConfigWithEnv(filename string) (*Config, error) {
	config, err := LoadConfig(filename)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if present
	if addr := os.Getenv("MULTIMINER_LISTEN_ADDRESS"); addr != "" {
		config.Server.ListenAddress = addr
	}

	if level := os.Getenv("MULTIMINER_LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}

	if timeoutStr := os.Getenv("MULTIMINER_PROBE_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Manager.ProbeTimeout = timeout
		}
	}

	if maxIdleStr := os.Getenv("MULTIMINER_MAX_IDLE_CONNECTIONS"); maxIdleStr != "" {
		if maxIdle, err := strconv.Atoi(maxIdleStr); err == nil {
			config.Pool.MaxIdleConnections = maxIdle
		}
	}

	if maxOpenStr := os.Getenv("MULTIMINER_MAX_OPEN_CONNECTIONS"); maxOpenStr != "" {
		if maxOpen, err := strconv.Atoi(maxOpenStr); err == nil {
			config.Pool.MaxOpenConnections = maxOpen
		}
	}

	return config, nil
}

// SaveConfig saves configuration to a JSON file
func (c *Config) SaveConfig(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ToManagerOptions converts Config to legacy ManagerOptions
func (c *Config) ToManagerOptions() ManagerOptions {
	return ManagerOptions{
		ProbeTimeout: c.Manager.ProbeTimeout,
	}
}

// GetLogLevel converts string log level to LogLevel
func (c *Config) GetLogLevel() LogLevel {
	switch c.Logging.Level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn", "warning":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

// Legacy function for backward compatibility
func defaultOptions() ManagerOptions {
	return ManagerOptions{ProbeTimeout: 1200 * time.Millisecond}
}
