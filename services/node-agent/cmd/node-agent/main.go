package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fregataa/aami/node-agent/internal/agent"
	"github.com/fregataa/aami/node-agent/internal/config"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	var (
		configPath     string
		bootstrapToken string
		showVersion    bool
	)

	flag.StringVar(&configPath, "config", "/etc/aami/agent.yaml", "Path to configuration file")
	flag.StringVar(&bootstrapToken, "bootstrap-token", "", "Bootstrap token for initial registration")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("aami-agent version %s (built %s)\n", version, buildTime)
		os.Exit(0)
	}

	// Setup logger
	logger := setupLogger()

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err, "path", configPath)
		os.Exit(1)
	}

	// Override bootstrap token from CLI if provided
	if bootstrapToken != "" {
		cfg.Agent.BootstrapToken = bootstrapToken
	}

	// Update logger with configured level
	logger = setupLoggerWithConfig(cfg.Logging)
	slog.SetDefault(logger)

	logger.Info("starting aami-agent",
		"version", version,
		"config", configPath,
	)

	// Create agent
	a, err := agent.New(cfg, logger)
	if err != nil {
		logger.Error("failed to create agent", "error", err)
		os.Exit(1)
	}

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("received signal, shutting down", "signal", sig)
		cancel()
	}()

	// Run agent
	if err := a.Run(ctx); err != nil {
		logger.Error("agent stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("agent stopped gracefully")
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func setupLoggerWithConfig(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
