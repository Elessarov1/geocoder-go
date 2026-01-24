package server

import (
	"fmt"
	"strings"
	"time"

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

	// shutdown_timeout (optional)
	var shutdownTimeout time.Duration
	if rawTO, ok := sec[keys.ShutdownTimeout]; ok && rawTO != nil {
		s, ok := rawTO.(string)
		if !ok {
			return nil, fmt.Errorf("%s.key %s must be duration string", keys.Server, keys.ShutdownTimeout)
		}
		d, err := time.ParseDuration(strings.TrimSpace(s))
		if err != nil {
			return nil, fmt.Errorf("%s.key %s invalid duration: %w", keys.Server, keys.ShutdownTimeout, err)
		}
		shutdownTimeout = d
	}

	// metrics (optional object)
	metrics := MetricsConfig{Enabled: false, Path: "/metrics"}
	if msec, exists, err := config.OptionalSection(sec, keys.Metrics); err != nil {
		return nil, fmt.Errorf("%s.%w", keys.Server, err)
	} else if exists {
		metrics.Enabled = config.OptionalBool(msec, keys.Enabled, false)
		metrics.Path = config.OptionalString(msec, keys.Path, "/metrics")
	}

	// swagger (optional object)
	swagger := SwaggerConfig{Enabled: false, YAMLPath: "/swagger.yaml", UIPath: "/swagger-ui/"}
	if ssec, exists, err := config.OptionalSection(sec, keys.Swagger); err != nil {
		return nil, fmt.Errorf("%s.%w", keys.Server, err)
	} else if exists {
		swagger.Enabled = config.OptionalBool(ssec, keys.Enabled, false)
		swagger.YAMLPath = config.OptionalString(ssec, keys.YAMLPath, "/swagger.yaml")
		swagger.UIPath = config.OptionalString(ssec, keys.UIPath, "/swagger-ui/")
	}

	// cors (optional object)
	cors := CORSConfig{Enabled: false}
	if csec, exists, err := config.OptionalSection(sec, keys.CORS); err != nil {
		return nil, fmt.Errorf("%s.%w", keys.Server, err)
	} else if exists {
		cors.Enabled = config.OptionalBool(csec, keys.Enabled, false)
	}

	comp := New(Config{
		Host:            host,
		Port:            port,
		ShutdownTimeout: shutdownTimeout,
		Metrics:         metrics,
		Swagger:         swagger,
		CORS:            cors,
	}, bootstrap)

	return &core.Descriptor{
		Comp:      comp,
		Enabled:   true,
		DependsOn: meta.DependsOn,
	}, nil
}
