package response

import (
	"net/http"

	"github.com/go-api-template/app/domain/erring"
)

type Error struct {
	Type    string `json:"type"              extensions:"x-order=0" example:"srn:error:invalid_params"`
	Code    string `json:"code"              extensions:"x-order=1" example:"delivery_address:postal_code:regex-must-match"`
	Message string `json:"message,omitempty" extensions:"x-order=2" example:"delivery_address.postal_code must be in a valid format"`
}

var errorToStatusCode = map[error]int{
	// Shared
	erring.ErrEventInvalid: http.StatusBadRequest,
}

func StatusCodeFromError(err error) int {
	if statusCode, ok := errorToStatusCode[err]; ok {
		return statusCode
	}

	return http.StatusNotImplemented
}
