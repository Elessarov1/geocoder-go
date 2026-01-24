package bootstrap

import (
	"context"

	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"
	"github.com/Elessarov1/geocoder-go/internal/server"
	http_server "github.com/Elessarov1/service-kit/component/server"
	kitcore "github.com/Elessarov1/service-kit/core"
)

func Registry(api *geocoder_api.Service) kitcore.Registry {
	return kitcore.NewRegistry(
		http_server.Module(httpServer(api)),
	)
}

func httpServer(api *geocoder_api.Service) http_server.Bootstrap {
	return func(ctx context.Context, cfg http_server.Config) (http_server.Runtime, error) {
		return server.NewServer(ctx, &cfg, api)
	}
}
