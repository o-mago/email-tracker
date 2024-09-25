package resource

import (
	"net/http"
)

type Resource string

const (
	// User.
	SrnResourceUser Resource = "srn:resource:user"

	// Error.
	SrnErrorBadRequest          Resource = "srn:error:invalid_params"
	SrnErrorUnauthorized        Resource = "srn:error:unauthorized"
	SrnErrorForbidden           Resource = "srn:error:forbidden"
	SrnErrorNotFound            Resource = "srn:error:resource_not_found"
	SrnErrorMethodNotAllowed    Resource = "srn:error:method_not_allowed"
	SrnErrorRequestTimeout      Resource = "srn:error:request_timeout"
	SrnErrorConflict            Resource = "srn:error:conflict"
	SrnErrorPreconditionFailed  Resource = "srn:error:precondition_failed"
	SrnErrorUnprocessableEntity Resource = "srn:error:unprocessable_entity"
	SrnErrorTooManyRequests     Resource = "srn:error:too_many_requests"
	SrnErrorServerError         Resource = "srn:error:server_error"
	SrnErrorNotImplemented      Resource = "srn:error:not_implemented"
	SrnErrorServiceUnavailable  Resource = "srn:error:service_unavailable"
)

var statusCodeToResource = map[int]Resource{
	http.StatusBadRequest:          SrnErrorBadRequest,
	http.StatusUnauthorized:        SrnErrorUnauthorized,
	http.StatusForbidden:           SrnErrorForbidden,
	http.StatusNotFound:            SrnErrorNotFound,
	http.StatusMethodNotAllowed:    SrnErrorMethodNotAllowed,
	http.StatusRequestTimeout:      SrnErrorRequestTimeout,
	http.StatusConflict:            SrnErrorConflict,
	http.StatusPreconditionFailed:  SrnErrorPreconditionFailed,
	http.StatusUnprocessableEntity: SrnErrorUnprocessableEntity,
	http.StatusTooManyRequests:     SrnErrorTooManyRequests,
	http.StatusInternalServerError: SrnErrorServerError,
	http.StatusNotImplemented:      SrnErrorNotImplemented,
	http.StatusServiceUnavailable:  SrnErrorServiceUnavailable,
}

//nolint:revive
func ResourceFromStatusCode(status int) Resource {
	return statusCodeToResource[status]
}
