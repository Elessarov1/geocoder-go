package server

import (
	"context"
	"net/http"

	"Geocoder/internal/server/oas"
)

// GET /geo/networks?isoCodes=RU&isoCodes=US
func (s *GeoCoderServer) GetCountryNetworks(_ context.Context, params oas.GetCountryNetworksParams) (oas.GetCountryNetworksRes, error) {
	if s.geo == nil {
		return nil, ErrResponse(http.StatusInternalServerError, "geoip.not_loaded", "geoip database is not loaded")
	}

	if len(params.IsoCodes) == 0 {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "isoCodes must not be empty")
	}

	out := make([]oas.IsoCodeNetworks, 0, len(params.IsoCodes))

	for _, iso := range params.IsoCodes {
		code := string(iso)

		ranges, ok := s.geo.RangesByCountryUnsafe(code)
		if !ok {
			return nil, ErrResponse(http.StatusNotFound, "geo.country_not_found", "unknown iso code: "+code)
		}

		// serialize CIDR list
		networks := make([]oas.Cidr, len(ranges))
		for i, p := range ranges {
			networks[i] = oas.Cidr(p.String())
		}

		out = append(out, oas.IsoCodeNetworks{
			Code:     oas.IsoCode(code),
			Networks: networks,
		})
	}

	okRes := oas.GetCountryNetworksOKApplicationJSON(out)
	return &okRes, nil
}
