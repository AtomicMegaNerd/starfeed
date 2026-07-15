package testutils

import (
	"log/slog"
	"testing"
	"time"

	"github.com/lmittmann/tint"
)

func TestLogger(t *testing.T) *slog.Logger {
	return slog.New(
		tint.NewTextHandler(
			t.Output(),
			&tint.Options{Level: slog.LevelDebug, TimeFormat: time.RFC3339},
		),
	)
}
