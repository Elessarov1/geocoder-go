package server

import "Geocoder/internal/server/oas"

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
