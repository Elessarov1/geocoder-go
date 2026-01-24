package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Raw map[string]any

func ReadYAML(path string) (Raw, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	expanded, err := ExpandEnv(m)
	if err != nil {
		return nil, err
	}

	out, ok := expanded.(map[string]any)
	if !ok {
		return nil, err
	}
	return out, nil
}

func (c Raw) Section(name string) (map[string]any, bool) {
	v, ok := c[name]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]any)
	return m, ok
}
