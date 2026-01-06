package server

import (
	"context"
	"github.com/Elessarov1/geocoder-go/internal/server/oas"
)

// GET /geo/networks?isoCodes=RU&isoCodes=US
func (s *GeoCoderServer) GetCountryNetworks(ctx context.Context, params oas.GetCountryNetworksParams) (oas.GetCountryNetworksRes, error) {
	isoCodes := make([]string, 0, len(params.IsoCodes))
	for _, iso := range params.IsoCodes {
		isoCodes = append(isoCodes, string(iso))
	}

	items, err := s.api.GetCountryNetworks(ctx, isoCodes)
	if err != nil {
		return nil, s.toOASError(ctx, err)
	}

	out := make([]oas.IsoCodeNetworks, 0, len(items))
	for _, it := range items {
		networks := make([]oas.Cidr, len(it.Networks))
		for i, p := range it.Networks {
			networks[i] = oas.Cidr(p.String())
		}

		out = append(out, oas.IsoCodeNetworks{
			Code:     oas.IsoCode(it.Code),
			Networks: networks,
		})
	}

	ok := oas.GetCountryNetworksOKApplicationJSON(out)
	return &ok, nil
}

// GET /geo/networks/paged?isoCode=RU&page=0&size=1000
func (s *GeoCoderServer) GetCountryNetworksPaged(ctx context.Context, params oas.GetCountryNetworksPagedParams) (oas.GetCountryNetworksPagedRes, error) {
	pageData, err := s.api.GetCountryNetworksPaged(
		ctx,
		string(params.IsoCode),
		int(params.Page),
		int(params.Size),
	)
	if err != nil {
		return nil, s.toOASError(ctx, err)
	}

	content := make([]oas.Cidr, len(pageData.Content))
	for i, p := range pageData.Content {
		content[i] = oas.Cidr(p.String())
	}

	resp := oas.PageDataString{
		Content:       content,
		TotalElements: int64(pageData.TotalElements),
		TotalPages:    int64(pageData.TotalPages),
		Size:          int64(pageData.Size),
		Page:          int64(pageData.Page),
	}

	return &resp, nil
}
