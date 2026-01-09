package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eslutz/forwardarr/internal/config"
	"github.com/eslutz/forwardarr/internal/qbit"
	"github.com/eslutz/forwardarr/internal/server"
	"github.com/eslutz/forwardarr/internal/sync"
	"github.com/eslutz/forwardarr/internal/webhook"
	_ "github.com/eslutz/forwardarr/pkg/version"
)

func main() {
	cfg := config.Load()
	setupLogging(cfg.LogLevel)

	startupRetryDelay, startupTimeout := normalizeStartupSettings(cfg)
	startupMaxAttempts := calculateMaxAttempts(startupRetryDelay, startupTimeout)

	slog.Info("starting forwardarr",
		"gluetun_port_file", cfg.GluetunPortFile,
		"qbit_addr", cfg.QbitAddr,
		"startup_retry_delay", startupRetryDelay,
		"startup_timeout", startupTimeout,
		"startup_max_attempts", startupMaxAttempts,
		"sync_interval", cfg.SyncInterval,
		"metrics_port", cfg.MetricsPort,
		"webhook_enabled", cfg.WebhookEnabled,
	)

	qbitClient, err := createQbitClientWithRetry(cfg, startupRetryDelay, startupTimeout, startupMaxAttempts)
	if err != nil {
		slog.Error("failed to create qBittorrent client", "error", err)
		os.Exit(1)
	}

	var webhookClient *webhook.Client
	if cfg.WebhookEnabled {
		webhookClient = webhook.NewClient(
			cfg.WebhookURL,
			cfg.WebhookTimeout,
			webhook.Template(cfg.WebhookTemplate),
			cfg.WebhookEvents,
		)
		slog.Info("webhook notifications enabled",
			"url", cfg.WebhookURL,
			"timeout", cfg.WebhookTimeout,
			"template", cfg.WebhookTemplate,
			"events", cfg.WebhookEvents,
		)
	}

	watcher, err := sync.NewWatcher(cfg.GluetunPortFile, qbitClient, webhookClient, cfg.SyncInterval)
	if err != nil {
		slog.Error("failed to create file watcher", "error", err)
		os.Exit(1)
	}

	srv := server.NewServer(cfg.MetricsPort, qbitClient)

	// Start HTTP server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start watcher in goroutine
	watcherDone := make(chan error, 1)
	go func() {
		watcherDone <- watcher.Start()
	}()

	// Wait for shutdown signal or watcher error
	select {
	case <-ctx.Done():
		slog.Info("received shutdown signal, gracefully stopping...")

		// Give time for cleanup
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}

		slog.Info("shutdown complete")

	case err := <-watcherDone:
		if err != nil {
			slog.Error("watcher failed", "error", err)
			os.Exit(1)
		}
	}
}

func setupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

func createQbitClientWithRetry(cfg *config.Config, retryDelay, startupTimeout time.Duration, maxAttempts int) (*qbit.Client, error) {
	startTime := time.Now()
	deadline := startTime.Add(startupTimeout)

	var lastErr error
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		slog.Info("connecting to qBittorrent",
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"qbit_addr", cfg.QbitAddr,
		)

		client, err := qbit.NewClient(cfg.QbitAddr, cfg.QbitUser, cfg.QbitPass)
		if err == nil {
			slog.Info("connected to qBittorrent",
				"attempt", attempt,
				"elapsed", time.Since(startTime),
			)
			return client, nil
		}

		lastErr = err
		remaining := max(time.Until(deadline), 0)
		shouldRetry := remaining > 0
		sleep := time.Duration(0)
		if shouldRetry {
			sleep = exponentialBackoffDelay(attempt, retryDelay, remaining)
		}

		logMsg := "qBittorrent connection failed"
		if shouldRetry {
			logMsg = "qBittorrent connection failed, will retry"
		}

		slog.Warn(logMsg,
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"qbit_addr", cfg.QbitAddr,
			"retry_delay", sleep,
			"remaining_timeout", remaining,
			"error", err,
		)

		if !shouldRetry || sleep <= 0 {
			break
		}

		time.Sleep(sleep)
	}

	return nil, fmt.Errorf("failed to connect to qBittorrent after %d attempts within %s: %w", attempt, startupTimeout, lastErr)
}

func normalizeStartupSettings(cfg *config.Config) (time.Duration, time.Duration) {
	retryDelay := cfg.StartupRetryDelay
	if retryDelay <= 0 {
		retryDelay = 5 * time.Second
	}

	timeout := cfg.StartupTimeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	return retryDelay, timeout
}
