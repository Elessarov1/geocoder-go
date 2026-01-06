package geoip

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

const UnknownISO = "ZZ"

type CountryInfo struct {
	ISOCode string `maxminddb:"iso_code"`
}

type Record struct {
	Country           CountryInfo `maxminddb:"country"`
	RegisteredCountry CountryInfo `maxminddb:"registered_country"`
}

type Options struct {
	SkipAliasedNetworks bool
	UnknownISO          string // fallback, "ZZ" code
}

func DefaultOptions() Options {
	return Options{SkipAliasedNetworks: true, UnknownISO: UnknownISO}
}

func Load(ctx context.Context, mmdbPath string, opt Options) (*Store, error) {
	if opt.UnknownISO == "" {
		opt.UnknownISO = UnknownISO
	}

	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		return nil, fmt.Errorf("open mmdb: %w", err)
	}
	defer db.Close()

	s := &Store{
		isoByID:   make([]string, 0, 256),
		idByISO:   make(map[string]CountryID, 256),
		byCountry: make([][]netip.Prefix, 0, 256),

		byV4: make(map[v4Key]CountryID, 600000),
		byV6: make(map[netip.Prefix]CountryID, 500000),
	}

	getOrCreateID := func(iso string) CountryID {
		if id, ok := s.idByISO[iso]; ok {
			return id
		}
		id := CountryID(len(s.isoByID))
		s.idByISO[iso] = id
		s.isoByID = append(s.isoByID, iso)
		s.byCountry = append(s.byCountry, nil)
		return id
	}

	var iter *maxminddb.Networks
	if opt.SkipAliasedNetworks {
		iter = db.Networks(maxminddb.SkipAliasedNetworks)
	} else {
		iter = db.Networks()
	}

	for iter.Next() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var rec Record
		ipNet, err := iter.Network(&rec)
		if err != nil {
			return nil, fmt.Errorf("iterate network: %w", err)
		}

		iso := normalizeISO(rec.Country.ISOCode)
		if iso == "" {
			iso = normalizeISO(rec.RegisteredCountry.ISOCode)
		}
		if iso == "" {
			iso = opt.UnknownISO
		}

		pfx, err := ipNetToPrefix(ipNet)
		if err != nil {
			return nil, fmt.Errorf("convert network %v: %w", ipNet, err)
		}

		pfx = pfx.Masked()
		id := getOrCreateID(iso)

		s.byCountry[id] = append(s.byCountry[id], pfx)

		s.stats.TotalNetworks++
		if pfx.Addr().Is4() {
			s.stats.V4Networks++
			s.byV4[makeV4Key(pfx)] = id
		} else {
			s.stats.V6Networks++
			s.byV6[pfx] = id
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	s.finalize()
	return s, nil
}

func normalizeISO(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.ToUpper(s)
}

func ipNetToPrefix(n *net.IPNet) (netip.Prefix, error) {
	if n == nil {
		return netip.Prefix{}, fmt.Errorf("nil ipnet")
	}
	ones, bits := n.Mask.Size()
	if ones < 0 || bits <= 0 {
		return netip.Prefix{}, fmt.Errorf("bad mask size")
	}
	ip := n.IP
	if ip == nil {
		return netip.Prefix{}, fmt.Errorf("nil ip")
	}

	if ip4 := ip.To4(); ip4 != nil {
		var a4 [4]byte
		copy(a4[:], ip4)
		return netip.PrefixFrom(netip.AddrFrom4(a4), ones), nil
	}

	ip16 := ip.To16()
	if ip16 == nil || len(ip16) != net.IPv6len {
		return netip.Prefix{}, fmt.Errorf("bad ipv6 ip: %v", ip)
	}
	var a16 [16]byte
	copy(a16[:], ip16)
	return netip.PrefixFrom(netip.AddrFrom16(a16), ones), nil
}
