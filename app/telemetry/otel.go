package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-api-template/app/config"
	"github.com/go-api-template/app/library/ctxkey"
)

var (
	_otelEnv     = "undefined"
	_otelVersion = "undefined"
)

type Otel struct {
	Tracer   trace.Tracer
	provider *sdktrace.TracerProvider
	exporter *otlptrace.Exporter
}

type SpanLink struct {
	TraceID [16]byte `json:"trace_id"`
	SpanID  [8]byte  `json:"span_id"`
}

func (t *Otel) Close(ctx context.Context) error {
	const operation = "Telemetry.Otel.Close"

	var errs error

	if err := t.provider.ForceFlush(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s -> provider flush: %w", operation, err))
	}

	if err := t.provider.Shutdown(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s -> provider shutdown: %w", operation, err))
	}

	if err := t.exporter.Shutdown(ctx); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%s -> exporter shutdown: %w", operation, err))
	}

	return errs
}

func NewOtel(ctx context.Context, cfg config.Otel, env, version string) (Otel, error) {
	const operation = "Telemetry.NewOtel"

	_otelEnv, _otelVersion = env, version

	resource, err := otelNewResource(cfg)
	if err != nil {
		return Otel{}, fmt.Errorf("%s -> new resource: %w", operation, err)
	}

	exporter, err := otlptrace.New(ctx, otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(cfg.CollectorEndpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(cfg.ExporterTimeout),
	))
	if err != nil {
		return Otel{}, fmt.Errorf("%s -> new exporter: %w", operation, err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithExportTimeout(cfg.ExporterTimeout)),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SamplingRatio)),
	)

	tracer := provider.Tracer(cfg.ServiceName)

	// Set global tracer provider & text propagators
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		b3.New(b3.WithInjectEncoding(b3.B3SingleHeader|b3.B3MultipleHeader)),
		propagation.TraceContext{},
		propagation.Baggage{},
		xray.Propagator{},
	))

	return Otel{
		tracer,
		provider,
		exporter,
	}, nil
}

// ContextWithTracer returns a new context derived from ctx that
// is associated with the given tracer.
func ContextWithTracer(parent context.Context, tracer trace.Tracer) context.Context {
	return context.WithValue(parent, tracerCtxKey{}, tracer)
}

// OtelAttrsFromCtx generates a list of otel attributes from context keys.
func OtelAttrsFromCtx(ctx context.Context) []attribute.KeyValue {
	attrs := otelDefaultAttrs()

	if requestID, ok := ctxkey.GetRequestID(ctx); ok {
		attrs = append(attrs, attribute.String("request_id", requestID))
	}

	if idempotencyKey, ok := ctxkey.GetIdempotencyKey(ctx); ok {
		attrs = append(attrs, attribute.String("idempotency_key", idempotencyKey))
	}

	return attrs
}

func otelDefaultAttrs() []attribute.KeyValue {
	// https://docs.datadoghq.com/tracing/trace_collection/tracing_naming_convention/#span-tag-naming-convention
	return []attribute.KeyValue{
		attribute.String("env", _otelEnv),
		attribute.String("version", _otelVersion),
		attribute.String("language", "go"),
		attribute.String("component", "telemetry.otel"),
	}
}

func otelNewResource(cfg config.Otel) (*resource.Resource, error) {
	const operation = "Telemetry.otelNewResource"

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.DeploymentEnvironmentKey.String(_otelEnv),
			semconv.ServiceInstanceIDKey.String(uuid.NewString()),
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceNamespaceKey.String(cfg.ServiceNamespace),
			semconv.ServiceVersionKey.String(_otelVersion),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s -> %w", operation, err)
	}

	return res, nil
}

// StartInternalSpan starts a new Span with kind trace.SpanKindInternal.
func StartInternalSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return startSpan(ctx, name, trace.SpanKindInternal)
}

// StartServerSpan starts a new Span with kind trace.SpanKindServer.
func StartServerSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return startSpan(ctx, name, trace.SpanKindServer, opts...)
}

// StartClientSpan starts a new Span with kind trace.SpanKindClient.
func StartClientSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return startSpan(ctx, name, trace.SpanKindClient)
}

// StartProducerSpan starts a new Span with kind trace.SpanKindProducer.
func StartProducerSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return startSpan(ctx, name, trace.SpanKindProducer)
}

// StartConsumerSpan starts a new Span with kind trace.SpanKindConsumer.
func StartConsumerSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return startSpan(ctx, name, trace.SpanKindConsumer)
}

// startSpan starts a new Span with specified name and kind.
func startSpan(ctx context.Context, name string, kind trace.SpanKind, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	opts = append(opts, trace.WithSpanKind(kind), trace.WithAttributes(otelDefaultAttrs()...))

	return tracerFromCtx(ctx).Start(ctx, name, opts...) //nolint:spancheck
}

type tracerCtxKey struct{}

func tracerFromCtx(ctx context.Context) trace.Tracer {
	if tracer, ok := ctx.Value(tracerCtxKey{}).(trace.Tracer); ok {
		return tracer
	}

	return nil
}
