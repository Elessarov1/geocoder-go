package geoip

import (
	"encoding/binary"
	"net/netip"
	"sort"
	"strings"
)

type CountryID uint16
type v4Key uint64

func makeV4Key(p netip.Prefix) v4Key {
	a := p.Addr().As4()
	ip := binary.BigEndian.Uint32(a[:])
	bits := uint64(uint8(p.Bits())) // 0..32
	return v4Key((bits << 32) | uint64(ip))
}

type Stats struct {
	TotalNetworks   int
	UniqueCountries int
	V4Networks      int
	V6Networks      int
}

type Store struct {
	isoByID []string
	idByISO map[string]CountryID

	byCountry [][]netip.Prefix

	byV4 map[v4Key]CountryID        // O(1) exact CIDR lookup for IPv4
	byV6 map[netip.Prefix]CountryID // O(1) exact CIDR lookup for IPv6

	stats Stats
}

func (s *Store) Stats() Stats {
	return s.stats
}

func (s *Store) RangesCountByCountry(iso string) int {
	iso = strings.ToUpper(strings.TrimSpace(iso))
	id, ok := s.idByISO[iso]
	if !ok {
		return 0
	}
	return len(s.byCountry[id])
}

func (s *Store) CountryCodes() []string {
	out := make([]string, len(s.isoByID))
	copy(out, s.isoByID)
	sort.Strings(out)
	return out
}

func (s *Store) RangesByCountry(iso string) []netip.Prefix {
	iso = strings.ToUpper(strings.TrimSpace(iso))
	id, ok := s.idByISO[iso]
	if !ok {
		return nil
	}
	src := s.byCountry[id]
	out := make([]netip.Prefix, len(src))
	copy(out, src)
	return out
}

func (s *Store) CountryByCIDR(cidr string) (string, bool) {
	p, err := netip.ParsePrefix(strings.TrimSpace(cidr))
	if err != nil {
		return "", false
	}
	p = p.Masked()

	var (
		id CountryID
		ok bool
	)

	if p.Addr().Is4() {
		id, ok = s.byV4[makeV4Key(p)]
	} else {
		id, ok = s.byV6[p]
	}
	if !ok {
		return "", false
	}
	return s.isoByID[id], true
}

func (s *Store) finalize() {
	for i := range s.byCountry {
		rs := s.byCountry[i]
		sort.Slice(rs, func(a, b int) bool {
			ai, aj := rs[a].Addr(), rs[b].Addr()
			if ai != aj {
				return ai.Less(aj)
			}
			return rs[a].Bits() < rs[b].Bits()
		})
		s.byCountry[i] = rs
	}
	s.stats.UniqueCountries = len(s.isoByID)
}

// RangesByCountryUnsafe возвращает внутренний слайс (не копировать, не модифицировать!)
func (s *Store) RangesByCountryUnsafe(iso string) ([]netip.Prefix, bool) {
	iso = strings.ToUpper(strings.TrimSpace(iso))
	id, ok := s.idByISO[iso]
	if !ok {
		return nil, false
	}
	return s.byCountry[id], true
}
