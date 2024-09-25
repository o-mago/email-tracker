package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-api-template/app/gateway/api/rest/response"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		const operation = "Http.Middleware.Recoverer"

		defer func(ctx context.Context) {
			if rec := recover(); rec != nil {
				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("%v", rec)
				}

				resp := response.InternalServerError(err)

				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(resp.Status)
				json.NewEncoder(rw).Encode(resp.Payload) //nolint:errcheck

				slog.ErrorContext(
					ctx,
					fmt.Sprintf("%s -> %v", operation, err),
					slog.Any("stack_trace", string(debug.Stack())),
				)
			}
		}(context.WithoutCancel(req.Context()))

		next.ServeHTTP(rw, req)
	})
}
