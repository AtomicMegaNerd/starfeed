package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

// This configures the logger for our application setting the level to debug
// if specified.
func buildLogger(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	return slog.New(
		tint.NewTextHandler(
			os.Stderr,
			&tint.Options{Level: level, TimeFormat: time.RFC3339},
		),
	)
}
