package telemetry

import (
	"context"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/go-api-template/app/config"
	"github.com/go-api-template/app/library/ctxkey"
)

func SetLogger(cfg config.Config, buildTime, buildCommit, buildTag string, attrs ...slog.Attr) {
	var handler *slogHandler

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}

	if cfg.Environment == config.EnvLocal && !isRunningInDockerContainer() {
		opts.Level = slog.LevelDebug

		handler = &slogHandler{
			handlers: []slog.Handler{
				slog.NewJSONHandler(logTempFile(cfg.App.Name), opts),
				slog.NewTextHandler(os.Stdout, opts),
			},
		}
	} else {
		handler = &slogHandler{
			handlers: []slog.Handler{slog.NewJSONHandler(os.Stdout, opts)},
		}
	}

	attrs = append(attrs, logDefaultAttrs(cfg, buildTime, buildCommit, buildTag)...)

	handlerWithAttrs := handler.WithAttrs(attrs)

	logger := slog.New(handlerWithAttrs)

	slog.SetDefault(logger)
}

type slogHandler struct {
	handlers []slog.Handler
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0)

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		traceID := sc.TraceID().String()
		spanID := sc.SpanID().String()

		attrs = append(attrs,
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),

			// https://docs.datadoghq.com/tracing/other_telemetry/connect_logs_and_traces/opentelemetry/?tab=go
			slog.String("dd.trace_id", logConvertToDatadog(traceID)),
			slog.String("dd.span_id", logConvertToDatadog(spanID)),
		)
	}

	if requestID, ok := ctxkey.GetRequestID(ctx); ok {
		attrs = append(attrs, slog.String("request_id", requestID))
	}

	if idempotencyKey, ok := ctxkey.GetIdempotencyKey(ctx); ok {
		attrs = append(attrs, slog.String("idempotency_key", idempotencyKey))
	}

	for i := range h.handlers {
		err := h.handlers[i].WithAttrs(attrs).Handle(ctx, record)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	for i := range h.handlers {
		h.handlers[i].WithAttrs(attrs)
	}

	return &slogHandler{h.handlers}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	for i := range h.handlers {
		h.handlers[i].WithGroup(name)
	}

	return &slogHandler{h.handlers}
}

func LogAttrsFromHTTP(req *http.Request, status int, route string, duration *time.Duration) slog.Attr {
	attrs := []any{
		slog.Int("status_code", status),
		slog.String("status_text", http.StatusText(status)),
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("route", route),
		slog.String("proto", req.Proto),
		slog.String("remote_addr", req.RemoteAddr),
	}

	if host := req.Header.Get("Host"); host != "" {
		attrs = append(attrs, slog.String("host", host))
	}

	if trueClientIP := req.Header.Get("True-Client-Ip"); trueClientIP != "" {
		attrs = append(attrs, slog.String("true_client_ip", trueClientIP))
	}

	if forwardedFor := req.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		attrs = append(attrs, slog.String("forwarded_for", forwardedFor))
	}

	if realIP := req.Header.Get("X-Real-Ip"); realIP != "" {
		attrs = append(attrs, slog.String("real_ip", realIP))
	}

	if origin := req.Header.Get("Origin"); origin != "" {
		attrs = append(attrs, slog.String("origin", origin))
	}

	if referer := req.Referer(); referer != "" {
		attrs = append(attrs, slog.String("referer", referer))
	}

	if userAgent := req.UserAgent(); userAgent != "" {
		attrs = append(attrs, slog.String("user_agent", userAgent))
	}

	if duration != nil {
		attrs = append(attrs, slog.Duration("duration", *duration))
	}

	return slog.Group("http", attrs...)
}

func logDefaultAttrs(cfg config.Config, buildTime, buildCommit, buildTag string) []slog.Attr {
	return []slog.Attr{
		slog.String("build_time", buildTime),
		slog.String("build_commit", buildCommit),
		slog.String("build_tag", buildTag),

		// https://docs.datadoghq.com/tracing/other_telemetry/connect_logs_and_traces/opentelemetry/?tab=go
		slog.String("dd.env", string(cfg.Environment)),
		slog.String("dd.service", cfg.Otel.ServiceName),
		slog.String("dd.version", buildTag),
	}
}

func logConvertToDatadog(id string) string {
	const maxLen = 16

	if len(id) < maxLen {
		return ""
	}

	if len(id) > maxLen {
		id = id[maxLen:]
	}

	intValue, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return ""
	}

	return strconv.FormatUint(intValue, 10)
}

func logTempFile(appName string) *os.File {
	logDirPath := "/tmp/log"

	if _, err := os.Stat(logDirPath); os.IsNotExist(err) {
		if err = os.Mkdir(logDirPath, fs.ModePerm); err != nil {
			log.Fatalf("failed to create log dir: %v", err)
		}
	}

	file, err := os.CreateTemp(logDirPath, appName+"_*.log")
	if err != nil {
		log.Fatalf("failed to create log file: %v", err)
	}

	return file
}

func isRunningInDockerContainer() bool {
	_, err := os.Stat("/.dockerenv")

	return err == nil
}
