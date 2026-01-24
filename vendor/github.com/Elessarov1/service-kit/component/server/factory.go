package server

import (
	"fmt"
	"github.com/Elessarov1/service-kit/config"
	"github.com/Elessarov1/service-kit/core"
	"github.com/Elessarov1/service-kit/keys"
)

func Build(cfg config.Raw, bootstrap Bootstrap) (*core.Descriptor, error) {
	sec, ok := cfg.Section(keys.Server)
	if !ok {
		return nil, nil
	}

	meta := config.ReadMeta(sec)
	if !meta.Enabled {
		return &core.Descriptor{Enabled: false, DependsOn: meta.DependsOn}, nil
	}

	host, err := config.RequireString(sec, keys.Host)
	if err != nil {
		return nil, fmt.Errorf("%s.%w", keys.Server, err)
	}

	port, err := config.RequireInt(sec, keys.Port)
	if err != nil {
		return nil, fmt.Errorf("%s.%w", keys.Server, err)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("%s.%s out of range: %d", keys.Server, keys.Port, port)
	}

	swagger := config.OptionalBool(sec, keys.Swagger, false)
	cors := config.OptionalBool(sec, keys.CorsEnabled, false)

	comp := New(Config{
		Host:        host,
		Port:        port,
		Swagger:     swagger,
		CorsEnabled: cors,
	}, bootstrap)

	return &core.Descriptor{
		Comp:      comp,
		Enabled:   true,
		DependsOn: meta.DependsOn,
	}, nil
}
