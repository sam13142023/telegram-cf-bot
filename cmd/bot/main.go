package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"telegram-cf-bot/internal/bot"
	"telegram-cf-bot/internal/config"
	"telegram-cf-bot/internal/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.Fatal("failed to load configuration: %v", err)
	}

	// Initialize logger
	logCfg := logger.Config{
		Level:    cfg.Logging.Level,
		ToFile:   cfg.Logging.ToFile,
		FilePath: cfg.Logging.FilePath,
	}

	if err := logger.Initialize(logCfg); err != nil {
		logger.Fatal("failed to initialize logger: %v", err)
	}

	logger.WithFields(map[string]interface{}{"version": "2.0.0"}).Info("starting application")

	// Create bot instance
	b, err := bot.New(cfg)
	if err != nil {
		logger.Fatal("failed to create bot: %v", err)
	}

	logger.Info("bot created successfully")

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start bot in a goroutine
	go func() {
		if err := b.Start(); err != nil {
			logger.Error("bot error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("shutdown signal received, initiating graceful shutdown")

	// Stop bot gracefully
	b.Stop()

	logger.Info("application stopped")
}
