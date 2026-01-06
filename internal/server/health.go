package server

import (
	"context"
	"github.com/Elessarov1/geocoder-go/internal/server/oas"
)

// Get service health.
// GET /health
func (s *GeoCoderServer) GetHealth(ctx context.Context) (*oas.Health, error) {
	h, err := s.api.Health(ctx)
	if err != nil {
		return nil, s.toOASError(ctx, err)
	}

	return &oas.Health{
		Uptime:  h.UptimeSeconds,
		Version: h.Version,
	}, nil
}
