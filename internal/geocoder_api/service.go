package geocoder_api

import (
	"context"
	"net/netip"
	"strings"
	"time"

	"github.com/Elessarov1/geocoder-go/internal/common/version"
	"github.com/Elessarov1/geocoder-go/internal/geoip"

	"github.com/oschwald/maxminddb-golang"
)

type Service struct {
	store     *geoip.Store
	mmdb      *maxminddb.Reader
	startTime time.Time
}

func NewService(store *geoip.Store, mmdb *maxminddb.Reader, startTime time.Time) *Service {
	return &Service{
		store:     store,
		mmdb:      mmdb,
		startTime: startTime,
	}
}

var _ API = (*Service)(nil)

func (s *Service) Health(_ context.Context) (Health, error) {
	return Health{
		UptimeSeconds: int(time.Since(s.startTime).Seconds()),
		Version:       version.Version(),
	}, nil
}

func (s *Service) GetCountries(_ context.Context) ([]CountryRangeData, error) {
	if s.store == nil {
		return nil, &InvalidArgumentError{Msg: "geoip store is not loaded"}
	}

	codes := s.store.CountryCodes()
	out := make([]CountryRangeData, 0, len(codes))
	for _, code := range codes {
		out = append(out, CountryRangeData{
			Code:        code,
			RangesCount: s.store.RangesCountByCountry(code),
		})
	}
	return out, nil
}

func (s *Service) GetIpData(_ context.Context, ips []string) ([]GeoIPData, error) {
	if s.mmdb == nil {
		return nil, &InvalidArgumentError{Msg: "mmdb reader is not initialized"}
	}
	if len(ips) == 0 {
		return nil, &InvalidArgumentError{Msg: "ips must not be empty"}
	}

	out := make([]GeoIPData, 0, len(ips))

	for _, ipStr := range ips {
		ipStr = strings.TrimSpace(ipStr)
		if ipStr == "" {
			return nil, &InvalidArgumentError{Msg: "empty ip"}
		}

		addr, err := netip.ParseAddr(ipStr)
		if err != nil {
			return nil, &InvalidArgumentError{Msg: "invalid ip: " + ipStr}
		}

		var rec geoip.Record
		_ = s.mmdb.Lookup(addr.AsSlice(), &rec)

		iso := normalizeISO(rec.Country.ISOCode)
		if iso == "" {
			iso = normalizeISO(rec.RegisteredCountry.ISOCode)
		}
		if iso == "" {
			iso = geoip.UnknownISO
		}

		out = append(out, GeoIPData{
			IP:   ipStr,
			Code: iso,
		})
	}

	return out, nil
}

func (s *Service) GetCountryNetworks(_ context.Context, isoCodes []string) ([]IsoCodeNetworks, error) {
	if s.store == nil {
		return nil, &InvalidArgumentError{Msg: "geoip store is not loaded"}
	}
	if len(isoCodes) == 0 {
		return nil, &InvalidArgumentError{Msg: "isoCodes must not be empty"}
	}

	out := make([]IsoCodeNetworks, 0, len(isoCodes))

	for _, code := range isoCodes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			return nil, &InvalidArgumentError{Msg: "isoCode must not be empty"}
		}

		ranges, ok := s.store.RangesByCountryUnsafe(code)
		if !ok {
			return nil, &NotFoundError{Msg: "unknown iso code: " + code}
		}

		out = append(out, IsoCodeNetworks{
			Code:     code,
			Networks: ranges, // read-only view, без копирования
		})
	}

	return out, nil
}

func (s *Service) GetCountryNetworksPaged(_ context.Context, isoCode string, page, size int) (PageData, error) {
	if s.store == nil {
		return PageData{}, &InvalidArgumentError{Msg: "geoip store is not loaded"}
	}

	isoCode = strings.ToUpper(strings.TrimSpace(isoCode))
	if isoCode == "" {
		return PageData{}, &InvalidArgumentError{Msg: "isoCode must not be empty"}
	}
	if page < 0 {
		return PageData{}, &InvalidArgumentError{Msg: "page must be >= 0"}
	}
	if size <= 0 {
		return PageData{}, &InvalidArgumentError{Msg: "size must be >= 1"}
	}

	ranges, ok := s.store.RangesByCountryUnsafe(isoCode)
	if !ok {
		return PageData{}, &NotFoundError{Msg: "unknown iso code: " + isoCode}
	}

	total := len(ranges)
	from := page * size
	if from > total {
		from = total
	}
	to := from + size
	if to > total {
		to = total
	}

	totalPages := (total + size - 1) / size

	return PageData{
		Content:       ranges[from:to], // slice view
		TotalElements: total,
		TotalPages:    totalPages,
		Page:          page,
		Size:          size,
	}, nil
}

func normalizeISO(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s)
}
