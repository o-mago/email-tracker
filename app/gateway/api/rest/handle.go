package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cep21/circuit/v4"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-api-template/app/config"
	"github.com/go-api-template/app/domain/erring"
	"github.com/go-api-template/app/gateway/api/rest/response"
	"github.com/go-api-template/app/library/logutil"
	"github.com/go-api-template/app/library/resource"
	"github.com/go-api-template/app/telemetry"
)

//nolint:nestif
func HandleWithCircuit(circ *circuit.Circuit, cuircuitCfg config.CircuitBreaker, cache cache, route string, handler func(*http.Request) *response.Response) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		span := trace.SpanFromContext(req.Context())
		span.SetAttributes(telemetry.OtelAttrsFromCtx(req.Context())...)

		code, desc := codes.Ok, ""

		var (
			resp  *response.Response
			start time.Time
		)

		err := circ.Execute(req.Context(), func(ctx context.Context) error {
			start = time.Now()

			resp = handler(req.WithContext(ctx))

			// If the error is expected, we don't want to open the circuit breaker,
			// so we return a circuit.SimpleBadRequest error.
			if errors.Is(resp.InternalErr, erring.ErrExpected) {
				// Also, if the circuit is already open, we close it again (the lib is not doing it by itself)
				circ.CloseCircuit(ctx)

				return circuit.SimpleBadRequest{Err: resp.InternalErr}
			}

			return resp.InternalErr
		}, nil)
		if err != nil {
			code, desc = codes.Error, err.Error()

			// If we have an error but not a response error, we can assume that
			// it is a circuit breaker error and we can replace the original
			// response.
			// When a circuit breaker timeout is triggered, the context reaches
			// its deadline and our application returns a context.DeadlineExceeded
			// error, so we need to handle it as a CB error.
			if resp == nil || resp.InternalErr == nil || errors.Is(err, context.DeadlineExceeded) {
				var spanLink telemetry.SpanLink

				_ = cache.Get(req.Context(), circ.Name(), &spanLink)

				// Creates a new span with a linked span from the last error that caused the circuit to open.
				// It makes it easier to locate the root cause of the error
				_, span = telemetry.StartServerSpan(req.Context(), "circuit-break-cause", trace.WithLinks(trace.Link{
					SpanContext: trace.NewSpanContext(trace.SpanContextConfig{
						TraceID: spanLink.TraceID,
						SpanID:  spanLink.SpanID,
					}),
				}))
				span.SetAttributes(telemetry.OtelAttrsFromCtx(req.Context())...)

				defer span.End()

				resp = handleCircuitBreakerErrorResponse(err)
			} else {
				// Stores the reference from the last error on our cache (trace_id and span_id), so we can retrieve
				// it's info to link to a possible circuit break error span
				// SpanLink: https://opentelemetry.io/docs/concepts/signals/traces/#span-links
				spanLink := telemetry.SpanLink{
					TraceID: span.SpanContext().TraceID(),
					SpanID:  span.SpanContext().SpanID(),
				}

				spanLinkTTL := cuircuitCfg.SleepWindow + 1*time.Second

				_ = cache.Set(req.Context(), circ.Name(), spanLink, spanLinkTTL)
			}

			// If the error is a circuit.SimpleBadRequest, it means it is an
			// expected error, so we want to handle the original error.
			var expected circuit.SimpleBadRequest

			asExpected := errors.As(err, &expected)
			if asExpected {
				code, desc, err = codes.Ok, "", expected.Cause()
			}

			logLevel := slog.LevelError
			if asExpected || resp.Status == http.StatusBadRequest {
				logLevel = slog.LevelWarn
			}

			if !resp.OmitLogs {
				logAttrs := logAttrs(req, resp, route, time.Since(start))
				slog.Log(req.Context(), logLevel, err.Error(), logAttrs...)
			}

			span.RecordError(err)
		}

		err = sendJSON(rw, resp.Status, resp.Payload, resp.Headers)
		if err != nil {
			code, desc = codes.Error, err.Error()
			span.RecordError(err)

			logAttrs := logAttrs(req, resp, route, time.Since(start))
			slog.ErrorContext(req.Context(), err.Error(), logAttrs...)
		}

		span.SetStatus(code, desc)
	}
}

func sendJSON(rw http.ResponseWriter, statusCode int, payload any, header map[string]string) error {
	for key, value := range header {
		rw.Header().Set(key, value)
	}

	if payload == nil {
		rw.WriteHeader(statusCode)

		return nil
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)

	err := json.NewEncoder(rw).Encode(payload)
	if err != nil {
		return fmt.Errorf("send json encode: %w", err)
	}

	return nil
}

func handleCircuitBreakerErrorResponse(err error) *response.Response {
	if errors.Is(err, context.DeadlineExceeded) {
		return &response.Response{
			Status: http.StatusRequestTimeout,
			Payload: response.Error{
				Type:    string(resource.SrnErrorRequestTimeout),
				Code:    "circuit-breaker:request-timeout",
				Message: "timeout",
			},
			InternalErr: err,
		}
	}

	var cbErr circuit.Error

	if errors.As(err, &cbErr) {
		switch {
		case cbErr.CircuitOpen():
			return &response.Response{
				Status: http.StatusServiceUnavailable,
				Payload: response.Error{
					Type:    string(resource.SrnErrorServiceUnavailable),
					Code:    "circuit-breaker:service-unavailable",
					Message: cbErr.Error(),
				},
				InternalErr: err,
			}
		case cbErr.ConcurrencyLimitReached():
			return &response.Response{
				Status: http.StatusTooManyRequests,
				Payload: response.Error{
					Type:    string(resource.SrnErrorTooManyRequests),
					Code:    "circuit-breaker:too-many-requests",
					Message: cbErr.Error(),
				},
				InternalErr: err,
			}
		}
	}

	return response.InternalServerError(err)
}

func logAttrs(req *http.Request, resp *response.Response, route string, duration time.Duration) []any {
	attrs := make([]any, 0)

	for key, value := range resp.LogAttrs {
		attrs = append(attrs, slog.Any(key, value))
	}

	logMetadata := logutil.WithMetadata(attrs...)
	logAttrs := telemetry.LogAttrsFromHTTP(req, resp.Status, route, &duration)

	return []any{logMetadata, logAttrs}
}

type cache interface {
	Get(ctx context.Context, key string, objByRef any) error
	Set(ctx context.Context, key string, obj any, ttl time.Duration) error
}
