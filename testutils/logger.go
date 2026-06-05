package testutils

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func TestLogger() *slog.Logger {
	return slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)
}
