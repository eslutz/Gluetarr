package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: &Config{
				GluetunPortFile: "/tmp/gluetun/forwarded_port",
				QbitAddr:        "http://localhost:8080",
				QbitUser:        "admin",
				QbitPass:        "adminadmin",
				SyncInterval:    60 * time.Second,
				MetricsPort:     "9090",
				LogLevel:        "info",
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"GLUETUN_PORT_FILE":       "/custom/path/port",
				"QBIT_ADDR":               "http://custom:9090",
				"TORRENT_CLIENT_USER":     "testuser",
				"TORRENT_CLIENT_PASSWORD": "testpass",
				"SYNC_INTERVAL":           "120",
				"METRICS_PORT":            "8080",
				"LOG_LEVEL":               "debug",
			},
			expected: &Config{
				GluetunPortFile: "/custom/path/port",
				QbitAddr:        "http://custom:9090",
				QbitUser:        "testuser",
				QbitPass:        "testpass",
				SyncInterval:    120 * time.Second,
				MetricsPort:     "8080",
				LogLevel:        "debug",
			},
		},
		{
			name: "partial custom values",
			envVars: map[string]string{
				"TORRENT_CLIENT_USER": "myuser",
				"LOG_LEVEL":           "warn",
			},
			expected: &Config{
				GluetunPortFile: "/tmp/gluetun/forwarded_port",
				QbitAddr:        "http://localhost:8080",
				QbitUser:        "myuser",
				QbitPass:        "adminadmin",
				SyncInterval:    60 * time.Second,
				MetricsPort:     "9090",
				LogLevel:        "warn",
			},
		},
		{
			name: "invalid sync interval defaults to 60s",
			envVars: map[string]string{
				"SYNC_INTERVAL": "invalid",
			},
			expected: &Config{
				GluetunPortFile: "/tmp/gluetun/forwarded_port",
				QbitAddr:        "http://localhost:8080",
				QbitUser:        "admin",
				QbitPass:        "adminadmin",
				SyncInterval:    60 * time.Second,
				MetricsPort:     "9090",
				LogLevel:        "info",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := Load()

			if cfg.GluetunPortFile != tt.expected.GluetunPortFile {
				t.Errorf("GluetunPortFile = %v, want %v", cfg.GluetunPortFile, tt.expected.GluetunPortFile)
			}
			if cfg.QbitAddr != tt.expected.QbitAddr {
				t.Errorf("QbitAddr = %v, want %v", cfg.QbitAddr, tt.expected.QbitAddr)
			}
			if cfg.QbitUser != tt.expected.QbitUser {
				t.Errorf("QbitUser = %v, want %v", cfg.QbitUser, tt.expected.QbitUser)
			}
			if cfg.QbitPass != tt.expected.QbitPass {
				t.Errorf("QbitPass = %v, want %v", cfg.QbitPass, tt.expected.QbitPass)
			}
			if cfg.SyncInterval != tt.expected.SyncInterval {
				t.Errorf("SyncInterval = %v, want %v", cfg.SyncInterval, tt.expected.SyncInterval)
			}
			if cfg.MetricsPort != tt.expected.MetricsPort {
				t.Errorf("MetricsPort = %v, want %v", cfg.MetricsPort, tt.expected.MetricsPort)
			}
			if cfg.LogLevel != tt.expected.LogLevel {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.expected.LogLevel)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns env value when set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "returns default when env not set",
			key:          "UNSET_KEY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "returns parsed duration when valid",
			key:          "TEST_DURATION",
			defaultValue: 60 * time.Second,
			envValue:     "120",
			expected:     120 * time.Second,
		},
		{
			name:         "returns default when env not set",
			key:          "UNSET_DURATION",
			defaultValue: 60 * time.Second,
			envValue:     "",
			expected:     60 * time.Second,
		},
		{
			name:         "returns default when env value is invalid",
			key:          "INVALID_DURATION",
			defaultValue: 60 * time.Second,
			envValue:     "not-a-number",
			expected:     60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
			}

			result := getDurationEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getDurationEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}
