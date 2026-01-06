package server

import (
	"Geocoder/internal/server/oas"
	"context"
	"net/http"
)

// GET /geo/countries
func (s *GeoCoderServer) GetCountries(ctx context.Context) (oas.GetCountriesRes, error) {
	if s.geo == nil {
		return nil, ErrResponse(http.StatusInternalServerError, "geoip.not_loaded", "geoip database is not loaded")
	}

	codes := s.geo.CountryCodes()
	out := make([]oas.CountryRangeData, 0, len(codes))

	for _, code := range codes {
		out = append(out, oas.CountryRangeData{
			Code:        oas.IsoCode(code),
			RangesCount: int32(s.geo.RangesCountByCountry(code)),
		})
	}

	ok := oas.GetCountriesOKApplicationJSON(out)
	return &ok, nil
}
