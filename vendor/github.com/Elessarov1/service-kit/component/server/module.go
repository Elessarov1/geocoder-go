package server

import (
	"github.com/Elessarov1/service-kit/config"
	"github.com/Elessarov1/service-kit/core"
)

// Module returns a core.Factory for the server component.
// Usage: kitserver.Module(bootstrap)
func Module(bootstrap Bootstrap) core.Factory {
	return func(cfg config.Raw) (*core.Descriptor, error) {
		return Build(cfg, bootstrap)
	}
}
