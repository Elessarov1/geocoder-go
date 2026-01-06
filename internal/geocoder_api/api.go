package geocoder_api

import (
	"context"
	"net/netip"
)

type Health struct {
	UptimeSeconds int
	Version       string
}

type CountryRangeData struct {
	Code        string
	RangesCount int
}

type GeoIPData struct {
	IP          string
	Code        string
	CountryName string
}

type IsoCodeNetworks struct {
	Code     string
	Networks []netip.Prefix
}

type PageData struct {
	Content       []netip.Prefix
	TotalElements int
	TotalPages    int
	Page          int
	Size          int
}

type API interface {
	Health(ctx context.Context) (Health, error)

	GetCountries(ctx context.Context) ([]CountryRangeData, error)
	GetIpData(ctx context.Context, ips []string) ([]GeoIPData, error)

	GetCountryNetworks(ctx context.Context, isoCodes []string) ([]IsoCodeNetworks, error)
	GetCountryNetworksPaged(ctx context.Context, isoCode string, page, size int) (PageData, error)
}
