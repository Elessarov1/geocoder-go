package bootstrap

import (
	"context"
	"net/http"

	Geocoder "github.com/Elessarov1/geocoder-go"
	"github.com/Elessarov1/geocoder-go/internal/common/logger"
	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"
	"github.com/Elessarov1/geocoder-go/internal/gprc_server"
	"github.com/Elessarov1/geocoder-go/internal/grpc/gen/geocoderv1"
	"github.com/Elessarov1/geocoder-go/internal/server"
	"github.com/Elessarov1/geocoder-go/internal/server/oas"
	kit_grpc "github.com/Elessarov1/service-kit/component/grpc"
	http_server "github.com/Elessarov1/service-kit/component/server"
	kitcore "github.com/Elessarov1/service-kit/core"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Registry(api *geocoder_api.Service) kitcore.Registry {
	return kitcore.NewRegistry(
		http_server.StdModule(httpServer(api)),
		kit_grpc.StdModule(grpcServer(api)),
	)
}

func httpServer(api *geocoder_api.Service) http_server.StdOptions {
	// We'll capture the service logger from ctx during Register().
	var lg *zap.SugaredLogger

	return http_server.StdOptions{
		Register: func(ctx context.Context, mux *http.ServeMux) error {
			// Capture logger once, derived from real service ctx.
			lg = logger.FromContext(ctx).Named("http").Sugar()

			// Build ogen handler and register routes.
			h := server.NewHandler(ctx, api)
			oasServer, err := oas.NewServer(h)
			if err != nil {
				return err
			}
			mux.Handle("/", oasServer)
			return nil
		},

		Swagger: http_server.SwaggerAssets{
			UIFS:     Geocoder.SwaggerUI,
			UISubdir: "_openapi/swaggerui",
			YAML:     Geocoder.Swagger,
		},

		Logger: func(msg string, kv ...any) {
			if lg != nil {
				lg.Infow(msg, kv...)
			}
		},
	}
}

func grpcServer(api geocoder_api.API) kit_grpc.StdOptions {
	var lg *zap.SugaredLogger

	return kit_grpc.StdOptions{
		Register: func(ctx context.Context, s *grpc.Server) error {
			lg = logger.FromContext(ctx).Named("grpc").Sugar()

			h := grpc_server.NewHandler(ctx, api)
			geocoderv1.RegisterGeocoderServiceServer(s, h)
			return nil
		},

		ServerOptions: []grpc.ServerOption{
			// Optional: interceptors, creds, etc.
			// grpc.UnaryInterceptor(unaryLoggingInterceptor(lg.Desugar())),
		},

		Logger: func(msg string, kv ...any) {
			if lg != nil {
				lg.Infow(msg, kv...)
			}
		},
	}
}
