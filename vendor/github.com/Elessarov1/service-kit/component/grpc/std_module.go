package grpc

import (
	"context"

	"github.com/Elessarov1/service-kit/config"
	"github.com/Elessarov1/service-kit/core"
)

func StdModule(opts StdOptions) core.Factory {
	return Module(func(ctx context.Context, cfg Config) (Runtime, error) {
		rt, err := NewStdGRPCRuntime(ctx, cfg, opts)
		if err != nil {
			return nil, err
		}
		return rt, nil
	})
}

func Module(bootstrap Bootstrap) core.Factory {
	return func(raw config.Raw) (*core.Descriptor, error) {
		return Build(raw, bootstrap)
	}
}
