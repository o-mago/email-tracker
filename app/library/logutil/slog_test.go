package logutil

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlog_WithMetadata(t *testing.T) {
	t.Parallel()

	logAttr := slog.Int("id", 123)
	logMetadata := WithMetadata(logAttr)

	assert.Equal(t, logMetadata, slog.Group("metadata", logAttr))
}
