package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Elessarov1/geocoder-go/internal/common/logger"
	"github.com/Elessarov1/geocoder-go/internal/grpc/gen/geocoderv1"

	"github.com/Elessarov1/geocoder-go/internal/config"
	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const shutdownTimeout = 10 * time.Second

type Server struct {
	geocoderv1.UnimplementedGeocoderServiceServer

	api geocoder_api.API
	lg  *zap.Logger

	grpc *grpc.Server
	lis  net.Listener
}

func New(ctx context.Context, cfg *config.GRPCConfig, api geocoder_api.API) (*Server, error) {
	lg := logger.FromContext(ctx).Named("grpc")

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen grpc %s: %w", addr, err)
	}

	s := &Server{
		api: api,
		lg:  lg,
		lis: lis,
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(unaryLoggingInterceptor(lg)),
	)
	s.grpc = grpcSrv

	geocoderv1.RegisterGeocoderServiceServer(grpcSrv, s)

	if cfg.Reflection {
		reflection.Register(grpcSrv)
	}

	return s, nil
}

func (s *Server) Serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		s.lg.Info("Shutting down gRPC server", zap.Duration("timeout", shutdownTimeout))

		done := make(chan struct{})
		go func() {
			s.grpc.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			s.lg.Info("gRPC server shutdown complete")
		case <-shutdownCtx.Done():
			s.lg.Warn("gRPC graceful stop timeout, forcing stop")
			s.grpc.Stop()
		}
	}()

	s.lg.Info("Starting gRPC server", zap.String("addr", s.lis.Addr().String()))
	return s.grpc.Serve(s.lis)
}

func (s *Server) Close() {
	if s.grpc != nil {
		s.grpc.Stop()
	}
	if s.lis != nil {
		_ = s.lis.Close()
	}
}

func toGRPCError(err error) error {
	var ia *geocoder_api.InvalidArgumentError
	if errors.As(err, &ia) {
		return status.Error(codes.InvalidArgument, ia.Error())
	}
	var nf *geocoder_api.NotFoundError
	if errors.As(err, &nf) {
		return status.Error(codes.NotFound, nf.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func (s *Server) GetHealth(ctx context.Context, _ *emptypb.Empty) (*geocoderv1.Health, error) {
	h, err := s.api.Health(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &geocoderv1.Health{
		UptimeSeconds: int32(h.UptimeSeconds),
		Version:       h.Version,
	}, nil
}

func (s *Server) GetCountries(ctx context.Context, _ *emptypb.Empty) (*geocoderv1.GetCountriesResponse, error) {
	items, err := s.api.GetCountries(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	out := make([]*geocoderv1.CountryRangeData, 0, len(items))
	for _, it := range items {
		out = append(out, &geocoderv1.CountryRangeData{
			Code:        it.Code,
			RangesCount: int32(it.RangesCount),
		})
	}
	return &geocoderv1.GetCountriesResponse{Countries: out}, nil
}

func (s *Server) GetIpData(ctx context.Context, req *geocoderv1.GetIpDataRequest) (*geocoderv1.GetIpDataResponse, error) {
	ips := make([]string, 0, len(req.GetIps()))
	for _, ip := range req.GetIps() {
		if ip.GetIp() != "" {
			ips = append(ips, ip.GetIp())
		}
	}

	items, err := s.api.GetIpData(ctx, ips)
	if err != nil {
		return nil, toGRPCError(err)
	}

	out := make([]*geocoderv1.GeoIpData, 0, len(items))
	for _, it := range items {
		out = append(out, &geocoderv1.GeoIpData{
			Ip:          it.IP,
			Code:        it.Code,
			CountryName: it.CountryName,
		})
	}
	return &geocoderv1.GetIpDataResponse{Items: out}, nil
}

func (s *Server) GetCountryNetworks(ctx context.Context, req *geocoderv1.GetCountryNetworksRequest) (*geocoderv1.GetCountryNetworksResponse, error) {
	items, err := s.api.GetCountryNetworks(ctx, req.GetIsoCodes())
	if err != nil {
		return nil, toGRPCError(err)
	}

	out := make([]*geocoderv1.IsoCodeNetworks, 0, len(items))
	for _, it := range items {
		nets := make([]string, len(it.Networks))
		for i, p := range it.Networks {
			nets[i] = p.String()
		}
		out = append(out, &geocoderv1.IsoCodeNetworks{
			Code:     it.Code,
			Networks: nets,
		})
	}
	return &geocoderv1.GetCountryNetworksResponse{Items: out}, nil
}

func (s *Server) GetCountryNetworksPaged(ctx context.Context, req *geocoderv1.GetCountryNetworksPagedRequest) (*geocoderv1.PageDataString, error) {
	pd, err := s.api.GetCountryNetworksPaged(ctx, req.GetIsoCode(), int(req.GetPage()), int(req.GetSize()))
	if err != nil {
		return nil, toGRPCError(err)
	}

	content := make([]string, len(pd.Content))
	for i, p := range pd.Content {
		content[i] = p.String()
	}

	return &geocoderv1.PageDataString{
		Content:       content,
		TotalElements: int64(pd.TotalElements),
		TotalPages:    int64(pd.TotalPages),
		Page:          int64(pd.Page),
		Size:          int64(pd.Size),
	}, nil
}
