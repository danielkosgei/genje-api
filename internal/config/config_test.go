package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test default values
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Aggregator.Interval != 30*time.Minute {
		t.Errorf("Expected default interval 30m, got %v", cfg.Aggregator.Interval)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{"TEST_VAR_1", "test_value", "default", "test_value"},
		{"TEST_VAR_2", "", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s, want %s", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	tests := []struct {
		key          string
		value        string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"TEST_DURATION_1", "5m", 10 * time.Minute, 5 * time.Minute},
		{"TEST_DURATION_2", "invalid", 10 * time.Minute, 10 * time.Minute},
		{"TEST_DURATION_3", "", 10 * time.Minute, 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getDuration(%s, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		key          string
		value        string
		defaultValue int
		expected     int
	}{
		{"TEST_INT_1", "42", 10, 42},
		{"TEST_INT_2", "invalid", 10, 10},
		{"TEST_INT_3", "", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if tt.value != "" {
				os.Setenv(tt.key, tt.value)
				defer os.Unsetenv(tt.key)
			}

			result := getInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getInt(%s, %d) = %d, want %d", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestConfigurationOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9000")
	os.Setenv("AGGREGATION_INTERVAL", "15m")
	os.Setenv("MAX_CONTENT_SIZE", "5000")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("AGGREGATION_INTERVAL")
		os.Unsetenv("MAX_CONTENT_SIZE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != "9000" {
		t.Errorf("Expected PORT override to 9000, got %s", cfg.Server.Port)
	}

	if cfg.Aggregator.Interval != 15*time.Minute {
		t.Errorf("Expected AGGREGATION_INTERVAL override to 15m, got %v", cfg.Aggregator.Interval)
	}

	if cfg.Aggregator.MaxContentSize != 5000 {
		t.Errorf("Expected MAX_CONTENT_SIZE override to 5000, got %d", cfg.Aggregator.MaxContentSize)
	}
}
