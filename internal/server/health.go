package server

import (
	"context"

	"github.com/Elessarov1/geocoder-go/internal/server/oas"
)

// Get service health.
// GET /health
func (h *GeoCoderHandler) GetHealth(ctx context.Context) (*oas.Health, error) {
	health, err := h.api.Health(ctx)
	if err != nil {
		return nil, h.toOASError(ctx, err)
	}

	return &oas.Health{
		Uptime:  health.UptimeSeconds,
		Version: health.Version,
	}, nil
}
