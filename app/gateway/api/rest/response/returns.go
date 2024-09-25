package response

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/go-api-template/app/domain/erring"
	"github.com/go-api-template/app/library/resource"
)

type Response struct { //nolint:errname
	Status      int
	Payload     any
	Headers     map[string]string
	InternalErr error
	LogAttrs    map[string]any
	OmitLogs    bool
}

func (r *Response) Error() string {
	return r.InternalErr.Error()
}

func (r *Response) WithHeaders(header map[string]string) *Response {
	r.Headers = header

	return r
}

func (r *Response) WithLogAttrs(attrs map[string]any) *Response {
	r.LogAttrs = attrs

	return r
}

func (r *Response) WithOmittedLogs() *Response {
	r.OmitLogs = true

	return r
}

// Success

func OK(payload any) *Response {
	return &Response{
		Status:  http.StatusOK,
		Payload: payload,
	}
}

func Created(payload any) *Response {
	return &Response{
		Status:  http.StatusCreated,
		Payload: payload,
	}
}

func Accepted(payload any) *Response {
	return &Response{
		Status:  http.StatusAccepted,
		Payload: payload,
	}
}

func NoContent() *Response {
	return &Response{
		Status: http.StatusNoContent,
	}
}

// Failure

func BadRequest(err error, message string) *Response {
	return &Response{
		Status:      http.StatusBadRequest,
		Payload:     makeBadRequestError(err, message),
		InternalErr: err,
	}
}

func Unauthorized() *Response {
	return &Response{
		Status: http.StatusUnauthorized,
		Payload: Error{
			Type:    string(resource.SrnErrorUnauthorized),
			Code:    "oops:unauthorized",
			Message: "user is not authorized to perform this operation",
		},
		InternalErr: errors.New("unauthorized"),
	}
}

func NotFound(err error, code, message string) *Response {
	return &Response{
		Status: http.StatusNotFound,
		Payload: Error{
			Type:    string(resource.SrnErrorNotFound),
			Code:    code,
			Message: message,
		},
		InternalErr: err,
	}
}

func MethodNotAllowed() Response {
	return Response{
		Status: http.StatusMethodNotAllowed,
		Payload: Error{
			Type:    string(resource.SrnErrorMethodNotAllowed),
			Code:    "oops:method-not-allowed",
			Message: "the http method used is not supported by this resource",
		},
		InternalErr: errors.New("method not allowed"),
	}
}

func InternalServerError(err error) *Response {
	return &Response{
		Status: http.StatusInternalServerError,
		Payload: Error{
			Type:    string(resource.SrnErrorServerError),
			Code:    "oops:internal-server-error",
			Message: "an unexpected error has occurred",
		},
		InternalErr: err,
	}
}

func AppError(err error) *Response {
	appError := new(erring.AppError)
	if errors.As(err, &appError) {
		status := StatusCodeFromError(appError)

		return &Response{
			Status: status,
			Payload: Error{
				Type:    string(resource.ResourceFromStatusCode(status)),
				Code:    appError.Code,
				Message: appError.Message,
			},
			InternalErr: err,
		}
	}

	return InternalServerError(err)
}

func AppExpectedError(err error) *Response {
	err = fmt.Errorf("%w: %w", erring.ErrExpected, err)

	return AppError(err)
}

func makeBadRequestError(err error, message string) Error {
	appError := new(erring.AppError)
	if errors.As(err, &appError) {
		return Error{
			Type:    string(resource.SrnErrorBadRequest),
			Code:    appError.Code,
			Message: message,
		}
	}

	return tryBuildValidationError("", err, message)
}

func tryBuildValidationError(prefix string, err error, message string) Error {
	var vErrs validation.Errors
	if errors.As(err, &vErrs) {
		for key, val := range vErrs {
			return tryBuildValidationError(fmt.Sprintf("%s:%s", prefix, key), val, message)
		}
	}

	var vErr validation.ErrorObject
	if errors.As(err, &vErr) {
		codePrefix := strings.TrimLeft(prefix, ":")
		msgPrefix := strings.ReplaceAll(codePrefix, ":", ".")

		return Error{
			Type:    string(resource.SrnErrorBadRequest),
			Code:    fmt.Sprintf("%s:%s", codePrefix, vErr.Code()),
			Message: fmt.Sprintf("%s %s", msgPrefix, vErr.Error()),
		}
	}

	return Error{
		Type:    string(resource.SrnErrorBadRequest),
		Code:    "oops:bad-request",
		Message: message,
	}
}
