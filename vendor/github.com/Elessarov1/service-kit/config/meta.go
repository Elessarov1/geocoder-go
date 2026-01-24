package config

import "github.com/Elessarov1/service-kit/keys"

type Meta struct {
	Enabled   bool
	DependsOn []string
}

func ReadMeta(sec map[string]any) Meta {
	return Meta{
		Enabled:   OptionalBool(sec, keys.Enabled, true),
		DependsOn: OptionalStringSlice(sec, keys.DependsOn),
	}
}
