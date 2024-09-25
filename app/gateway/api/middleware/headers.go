package middleware

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/go-api-template/app/library/ctxkey"
)

const (
	_authorizationHeaderName = "authorization"
	_requestIDHeaderName     = "x-request-id"
)

// HeadersToContext apply HTTP headers value to the context.
// Copies request id, idempotency key and authorization key to context.
func HeadersToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// Copy the authorization header value to the context.
		if authorization := req.Header.Get(_authorizationHeaderName); authorization != "" {
			ctx = ctxkey.PutAuthorizationHeader(ctx, authorization)
		}

		// Copy the request id header value to the context.
		requestID := req.Header.Get(_requestIDHeaderName)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx = ctxkey.PutRequestID(ctx, requestID)
		rw.Header().Set(_requestIDHeaderName, requestID)

		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
