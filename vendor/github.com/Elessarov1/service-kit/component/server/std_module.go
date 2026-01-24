package server

import (
	"context"

	"github.com/Elessarov1/service-kit/core"
)

// StdModule builds a server component using the standard net/http runtime.
func StdModule(opts StdOptions) core.Factory {
	return Module(func(ctx context.Context, cfg Config) (Runtime, error) {
		// Build standard runtime from resolved YAML config.
		return NewStdHTTPRuntime(ctx, cfg, opts)
	})
}
