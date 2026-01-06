package server

import (
	"Geocoder/internal/common/version"
	"Geocoder/internal/server/oas"
	"context"
	"time"
)

// Get service health.
// GET /health
func (s *GeoCoderServer) GetHealth(context.Context) (*oas.Health, error) {
	uptime := time.Since(s.startTime).Seconds()
	health := &oas.Health{
		Uptime:  int(uptime),
		Version: version.Version(),
	}
	return health, nil
}
