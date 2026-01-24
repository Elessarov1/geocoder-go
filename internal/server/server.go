package server

import (
	"context"

	"github.com/Elessarov1/geocoder-go/internal/common/logger"
	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"
	"github.com/Elessarov1/geocoder-go/internal/server/oas"

	"net/http"
	"time"

	"go.uber.org/zap"
)

type GeoCoderHandler struct {
	oas.UnimplementedHandler

	startTime time.Time
	lg        *zap.Logger

	api geocoder_api.API
}

var _ oas.Handler = (*GeoCoderHandler)(nil)

func NewHandler(ctx context.Context, api *geocoder_api.Service) *GeoCoderHandler {
	lg := logger.FromContext(ctx).Named("http")
	return &GeoCoderHandler{
		lg:        lg,
		startTime: time.Now(),
		api:       api,
	}
}

func (h *GeoCoderHandler) NewError(_ context.Context, err error) *oas.DefaultErrorStatusCode {
	h.lg.Error("API request error", zap.Error(err))
	return ErrResponse(http.StatusInternalServerError, "internal.error", err.Error())
}
