package server

import (
	"context"
	"net/http"

	"Geocoder/internal/server/oas"
)

// GET /geo/networks/paged?isoCode=RU&page=0&size=1000
func (s *GeoCoderServer) GetCountryNetworksPaged(_ context.Context, params oas.GetCountryNetworksPagedParams) (oas.GetCountryNetworksPagedRes, error) {
	if s.geo == nil {
		return nil, ErrResponse(http.StatusInternalServerError, "geoip.not_loaded", "geoip database is not loaded")
	}

	code := string(params.IsoCode)

	if params.Size <= 0 {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "size must be >= 1")
	}
	if params.Page < 0 {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "page must be >= 0")
	}

	ranges, ok := s.geo.RangesByCountryUnsafe(code)
	if !ok {
		return nil, ErrResponse(http.StatusNotFound, "geo.country_not_found", "unknown iso code: "+code)
	}

	total := len(ranges)
	size := int(params.Size)
	page := int(params.Page)

	from := page * size
	if from > total {
		from = total
	}
	to := from + size
	if to > total {
		to = total
	}

	// content
	slice := ranges[from:to]
	content := make([]oas.Cidr, len(slice))
	for i, p := range slice {
		content[i] = oas.Cidr(p.String())
	}

	// totals
	var totalPages int64
	if size > 0 {
		totalPages = int64((total + size - 1) / size)
	}

	resp := oas.PageDataString{
		Content:       content,
		TotalElements: int64(total),
		TotalPages:    totalPages,
		Size:          int64(size),
		Page:          int64(page),
	}

	okRes := resp
	return &okRes, nil
}
