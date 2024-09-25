package logutil

import "log/slog"

func WithMetadata(logAttrs ...any) slog.Attr {
	return slog.Group("metadata", logAttrs...)
}
