package mocks

import (
	"io"
	"log/slog"
)

func TestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
