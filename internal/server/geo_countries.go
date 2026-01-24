package server

import (
	"context"

	"github.com/Elessarov1/geocoder-go/internal/server/oas"
)

// GET /geo/countries
func (h *GeoCoderHandler) GetCountries(ctx context.Context) (oas.GetCountriesRes, error) {
	items, err := h.api.GetCountries(ctx)
	if err != nil {
		return nil, h.toOASError(ctx, err)
	}

	out := make([]oas.CountryRangeData, 0, len(items))
	for _, it := range items {
		out = append(out, oas.CountryRangeData{
			Code:        oas.IsoCode(it.Code),
			RangesCount: int32(it.RangesCount),
		})
	}

	ok := oas.GetCountriesOKApplicationJSON(out)
	return &ok, nil
}
