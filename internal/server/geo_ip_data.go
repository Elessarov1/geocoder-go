package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/Elessarov1/geocoder-go/internal/server/oas"
)

// POST /geo/ip_data
func (s *GeoCoderServer) GetIpData(ctx context.Context, req *oas.GeoPayload) (oas.GetIpDataRes, error) {
	if req == nil {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "request body is required")
	}

	ips := make([]string, 0, len(req.Ips))
	for _, item := range req.Ips {
		ipStr := strings.TrimSpace(string(item.IP))
		if ipStr != "" {
			ips = append(ips, ipStr)
		}
	}

	if len(ips) == 0 {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "ips must not be empty")
	}

	items, err := s.api.GetIpData(ctx, ips)
	if err != nil {
		return nil, s.toOASError(ctx, err)
	}

	out := make([]oas.GeoIpData, 0, len(items))
	for _, it := range items {
		out = append(out, oas.GeoIpData{
			IP:   oas.IpAddress(it.IP),
			Code: oas.IsoCode(it.Code),
		})
	}

	ok := oas.GetIpDataOKApplicationJSON(out)
	return &ok, nil
}
