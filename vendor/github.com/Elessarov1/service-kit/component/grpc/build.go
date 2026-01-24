package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Elessarov1/service-kit/config"
	"github.com/Elessarov1/service-kit/core"
	"github.com/Elessarov1/service-kit/keys"
)

type Bootstrap func(ctx context.Context, cfg Config) (Runtime, error)

func Build(raw config.Raw, bootstrap Bootstrap) (*core.Descriptor, error) {
	sec, ok := raw.Section(keys.GRPC)
	if !ok {
		return nil, nil
	}

	meta := config.ReadMeta(sec)
	if !meta.Enabled {
		return &core.Descriptor{Enabled: false, DependsOn: meta.DependsOn}, nil
	}

	host, err := config.RequireString(sec, keys.Host)
	if err != nil {
		return nil, fmt.Errorf("%s.%w", keys.GRPC, err)
	}

	port, err := config.RequireInt(sec, keys.Port)
	if err != nil {
		return nil, fmt.Errorf("%s.%w", keys.GRPC, err)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("%s.%s out of range: %d", keys.GRPC, keys.Port, port)
	}

	// shutdown_timeout: optional duration string
	shutdownTimeout := time.Duration(0)
	if v, ok := sec[keys.ShutdownTimeout]; ok {
		d, derr := config.AsDuration(v)
		if derr != nil {
			return nil, fmt.Errorf("%s.key %s: %w", keys.GRPC, keys.ShutdownTimeout, derr)
		}
		shutdownTimeout = d
	}

	// reflection.enabled: optional
	refEnabled := false
	if m, ok, err := config.OptionalMap(sec, keys.Reflection); err != nil {
		return nil, fmt.Errorf("%s.key %s: %w", keys.GRPC, keys.Reflection, err)
	} else if ok {
		refEnabled = config.OptionalBool(m, keys.Enabled, false)
	}

	comp := New(Config{
		Host:            host,
		Port:            port,
		ShutdownTimeout: shutdownTimeout,
		Reflection:      ReflectionConfig{Enabled: refEnabled},
	}, bootstrap)

	return &core.Descriptor{
		Comp:      comp,
		Enabled:   true,
		DependsOn: meta.DependsOn,
	}, nil
}
