package cmd

import (
	"context"
	"fmt"

	"github.com/Elessarov1/geocoder-go/internal/common/logger"
	"github.com/Elessarov1/geocoder-go/internal/config"

	"github.com/caarlos0/env/v11"
	"github.com/creasty/defaults"
)

func ReadConfig(ctx context.Context) (context.Context, config.Config, error) {
	var cfg config.Config

	// Create logger
	lg, err := logger.NewLogger()
	if err != nil {
		return ctx, cfg, fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer lg.Sync()
	ctx = logger.WithLogger(ctx, lg)

	// Load config
	if err := defaults.Set(&cfg); err != nil {
		return ctx, cfg, fmt.Errorf("failed to set defaults: %w", err)
	}
	if err := env.Parse(&cfg); err != nil {
		return ctx, cfg, fmt.Errorf("failed to parse env: %w", err)
	}

	// Set DEBUG level
	if cfg.GeoCoder.Debug {
		logger.SetDebug()
	}

	if err := cfg.Validate(); err != nil {
		return ctx, cfg, fmt.Errorf("config validation failed: %w", err)
	}

	return ctx, cfg, nil
}
