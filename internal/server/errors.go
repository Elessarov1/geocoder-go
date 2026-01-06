package server

import (
	"context"
	"github.com/Elessarov1/geocoder-go/internal/geocoder_api"
	"github.com/Elessarov1/geocoder-go/internal/server/oas"
	"net/http"

	"github.com/go-faster/errors"
	"go.uber.org/zap"
)

func ErrResponse(status int, code, desc string) *oas.DefaultErrorStatusCode {
	return &oas.DefaultErrorStatusCode{
		StatusCode: status,
		Response: oas.ErrorResponse{
			Result:  "ERROR",
			Content: oas.OptErrorResponseContent{},
			Error: oas.ErrorResponseError{
				Code:        code,
				Description: desc,
			},
		},
	}
}

// toOASError конвертит ошибки usecase-слоя в типизированную ogen-ошибку.
func (s *GeoCoderServer) toOASError(_ context.Context, err error) *oas.DefaultErrorStatusCode {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ErrResponse(http.StatusInternalServerError, "internal.canceled", err.Error())
	}

	var ia *geocoder_api.InvalidArgumentError
	if errors.As(err, &ia) {
		return ErrResponse(http.StatusBadRequest, "geo.bad_request", ia.Error())
	}

	var nf *geocoder_api.NotFoundError
	if errors.As(err, &nf) {
		return ErrResponse(http.StatusNotFound, "geo.not_found", nf.Error())
	}

	if s.lg != nil {
		s.lg.Error("API error", zap.Error(err))
	}

	return ErrResponse(http.StatusInternalServerError, "internal.error", err.Error())
}
