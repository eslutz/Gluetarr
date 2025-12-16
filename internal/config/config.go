package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	GluetunPortFile string
	QbitAddr        string
	QbitUser        string
	QbitPass        string
	SyncInterval    time.Duration
	MetricsPort     string
	LogLevel        string
}

func Load() *Config {
	return &Config{
		GluetunPortFile: getEnv("GLUETUN_PORT_FILE", "/tmp/gluetun/forwarded_port"),
		QbitAddr:        getEnv("TORRENT_CLIENT_URL", "http://localhost:8080"),
		QbitUser:        getEnv("TORRENT_CLIENT_USER", "admin"),
		QbitPass:        getEnv("TORRENT_CLIENT_PASSWORD", "adminadmin"),
		SyncInterval:    getDurationEnv("SYNC_INTERVAL", 5*time.Minute),
		MetricsPort:     getEnv("METRICS_PORT", "9090"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}
