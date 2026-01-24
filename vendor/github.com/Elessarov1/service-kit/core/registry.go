package core

import (
	"github.com/Elessarov1/service-kit/config"
)

type Descriptor struct {
	Comp      Component
	Enabled   bool
	DependsOn []string
	Index     int // порядок регистрации в registry (для стабильности)
}

type Factory func(cfg config.Raw) (*Descriptor, error)

type Registry []Factory

func BuildComponents(cfg config.Raw, reg Registry) ([]Component, error) {
	descs := make([]*Descriptor, 0, len(reg))

	for i, f := range reg {
		d, err := f(cfg)
		if err != nil {
			return nil, err
		}
		if d == nil {
			continue
		}
		d.Index = i
		if !d.Enabled {
			continue
		}
		descs = append(descs, d)
	}

	sorted, err := SortDescriptors(descs)
	if err != nil {
		return nil, err
	}

	out := make([]Component, 0, len(sorted))
	for _, d := range sorted {
		out = append(out, d.Comp)
	}
	return out, nil
}
