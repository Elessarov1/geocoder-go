package server

import (
	"Geocoder/internal/geoip"
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"

	"Geocoder/internal/server/oas"
)

type mmdbRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	RegisteredCountry struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"registered_country"`
}

func normISO(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s)
}

// POST /geo/ip_data
func (s *GeoCoderServer) GetIpData(ctx context.Context, req *oas.GeoPayload) (oas.GetIpDataRes, error) {
	if req == nil {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "request body is required")
	}

	if len(req.Ips) == 0 {
		return nil, ErrResponse(http.StatusBadRequest, "geo.bad_request", "ips must not be empty")
	}

	out := make([]oas.GeoIpData, 0, len(req.Ips))

	for _, item := range req.Ips {
		// имя поля может быть Ip или IP
		ipStr := strings.TrimSpace(string(item.IP))
		if ipStr == "" {
			return nil, ErrResponse(http.StatusBadRequest, "geo.bad_ip", "empty ip")
		}

		addr, err := netip.ParseAddr(ipStr)
		if err != nil {
			return nil, ErrResponse(http.StatusBadRequest, "geo.bad_ip", "invalid ip: "+ipStr)
		}

		var rec mmdbRecord
		_ = s.mmdb.Lookup(net.IP(addr.AsSlice()), &rec)

		iso := normISO(rec.Country.ISOCode)
		if iso == "" {
			iso = normISO(rec.RegisteredCountry.ISOCode)
		}
		if iso == "" {
			iso = geoip.UnknownISO
		}

		out = append(out, oas.GeoIpData{
			IP:   item.IP,
			Code: oas.IsoCode(iso),
		})
	}

	ok := oas.GetIpDataOKApplicationJSON(out)
	return &ok, nil
}
