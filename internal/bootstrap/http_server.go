package bootstrap

import (
	"context"

	"github.com/Elessarov1/geocoder-go/internal/config"
	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"
	"github.com/Elessarov1/geocoder-go/internal/server"
	kitserver "github.com/Elessarov1/service-kit/component/server"
)

func HTTPServerBootstrap(api *geocoder_api.Service) kitserver.Bootstrap {
	return func(ctx context.Context, sc kitserver.Config) (kitserver.Runtime, error) {
		srvCfg := &config.ServerConfig{
			Host:        sc.Host,
			Port:        sc.Port,
			Swagger:     sc.Swagger,
			CorsEnabled: sc.CorsEnabled,
		}
		return server.NewServer(ctx, srvCfg, api)
	}
}
