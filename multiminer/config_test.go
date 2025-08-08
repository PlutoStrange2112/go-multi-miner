package multiminer

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Server.ListenAddress != ":8080" {
		t.Errorf("Expected default listen address :8080, got %s", config.Server.ListenAddress)
	}

	if config.Manager.ProbeTimeout != 1200*time.Millisecond {
		t.Errorf("Expected probe timeout 1200ms, got %v", config.Manager.ProbeTimeout)
	}

	if !config.Security.EnableValidation {
		t.Error("Expected validation to be enabled by default")
	}

	if len(config.Security.AllowedPorts) == 0 {
		t.Error("Expected some allowed ports by default")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	config := DefaultConfig()
	config.Server.ListenAddress = ":9999"
	config.Logging.Level = "debug"

	// Save to temporary file
	tmpFile := "test_config.json"
	defer os.Remove(tmpFile)

	err := config.SaveConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config back
	loadedConfig, err := LoadConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.Server.ListenAddress != ":9999" {
		t.Errorf("Expected loaded listen address :9999, got %s", loadedConfig.Server.ListenAddress)
	}

	if loadedConfig.Logging.Level != "debug" {
		t.Errorf("Expected loaded log level debug, got %s", loadedConfig.Logging.Level)
	}
}

func TestConfigEnvironmentOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("MULTIMINER_LISTEN_ADDRESS", ":7777")
	os.Setenv("MULTIMINER_LOG_LEVEL", "warn")
	defer func() {
		os.Unsetenv("MULTIMINER_LISTEN_ADDRESS")
		os.Unsetenv("MULTIMINER_LOG_LEVEL")
	}()

	// Create temporary config file
	tmpFile := "test_env_config.json"
	defer os.Remove(tmpFile)

	config := DefaultConfig()
	config.SaveConfig(tmpFile)

	// Load with environment overrides
	loadedConfig, err := LoadConfigWithEnv(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config with env: %v", err)
	}

	if loadedConfig.Server.ListenAddress != ":7777" {
		t.Errorf("Expected env override address :7777, got %s", loadedConfig.Server.ListenAddress)
	}

	if loadedConfig.Logging.Level != "warn" {
		t.Errorf("Expected env override log level warn, got %s", loadedConfig.Logging.Level)
	}
}

func TestConfigLogLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn},
		{"error", LogLevelError},
		{"invalid", LogLevelInfo}, // default fallback
	}

	for _, test := range tests {
		config := DefaultConfig()
		config.Logging.Level = test.level

		if config.GetLogLevel() != test.expected {
			t.Errorf("Level %q: expected %v, got %v", test.level, test.expected, config.GetLogLevel())
		}
	}
}
