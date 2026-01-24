package grpc_server

import (
	"context"
	"errors"

	"github.com/Elessarov1/geocoder-go/internal/common/logger"
	"github.com/Elessarov1/geocoder-go/internal/grpc/gen/geocoderv1"

	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	geocoderv1.UnimplementedGeocoderServiceServer

	api geocoder_api.API
	lg  *zap.Logger
}

func NewHandler(ctx context.Context, api geocoder_api.API) *Handler {
	return &Handler{
		api: api,
		lg:  logger.FromContext(ctx).Named("grpc"),
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

func (h *Handler) GetHealth(ctx context.Context, _ *emptypb.Empty) (*geocoderv1.Health, error) {
	health, err := h.api.Health(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &geocoderv1.Health{
		UptimeSeconds: int32(health.UptimeSeconds),
		Version:       health.Version,
	}, nil
}

func (h *Handler) GetCountries(ctx context.Context, _ *emptypb.Empty) (*geocoderv1.GetCountriesResponse, error) {
	items, err := h.api.GetCountries(ctx)
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

func (h *Handler) GetIpData(ctx context.Context, req *geocoderv1.GetIpDataRequest) (*geocoderv1.GetIpDataResponse, error) {
	ips := make([]string, 0, len(req.GetIps()))
	for _, ip := range req.GetIps() {
		if ip.GetIp() != "" {
			ips = append(ips, ip.GetIp())
		}
	}

	items, err := h.api.GetIpData(ctx, ips)
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

func (h *Handler) GetCountryNetworks(ctx context.Context, req *geocoderv1.GetCountryNetworksRequest) (*geocoderv1.GetCountryNetworksResponse, error) {
	items, err := h.api.GetCountryNetworks(ctx, req.GetIsoCodes())
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

func (h *Handler) GetCountryNetworksPaged(ctx context.Context, req *geocoderv1.GetCountryNetworksPagedRequest) (*geocoderv1.PageDataString, error) {
	pd, err := h.api.GetCountryNetworksPaged(ctx, req.GetIsoCode(), int(req.GetPage()), int(req.GetSize()))
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

func (h *Handler) GetCountryNetworksStream(
	req *geocoderv1.GetCountryNetworksStreamRequest,
	stream geocoderv1.GeocoderService_GetCountryNetworksStreamServer,
) error {
	ctx := stream.Context()

	isoCodes := req.GetIsoCodes()
	if len(isoCodes) == 0 {
		return status.Error(codes.InvalidArgument, "iso_codes must not be empty")
	}

	chunkSize := int(req.GetChunkSize())
	if chunkSize <= 0 {
		chunkSize = 5000 // default
	}
	if chunkSize > 100000 {
		chunkSize = 100000
	}

	for _, code := range isoCodes {
		page := 0
		for {
			pd, err := h.api.GetCountryNetworksPaged(ctx, code, page, chunkSize)
			if err != nil {
				return toGRPCError(err)
			}

			nets := make([]string, len(pd.Content))
			for i, p := range pd.Content {
				nets[i] = p.String()
			}

			totalPages := pd.TotalPages
			last := false

			if totalPages <= 0 {
				totalPages = 1
			}
			if page >= totalPages-1 {
				last = true
			}

			if err := stream.Send(&geocoderv1.CountryNetworksChunk{
				Code:       code,
				Networks:   nets,
				Page:       int32(page),
				TotalPages: int32(totalPages),
				Last:       last,
			}); err != nil {
				return err
			}

			if last {
				break
			}
			page++
		}
	}

	return nil
}
