package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/runner"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("***********************************************")
	slog.Info(" Welcome to Starfeed")
	slog.Info("***********************************************")

	cfg, err := config.NewConfig(config.OSEnvGetter{})
	if err != nil {
		slog.Error("Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	handler := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if cfg.DebugMode {
		slog.Info("Debug mode enabled")
		handler = &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, handler))
	slog.SetDefault(logger)

	// In this case both os.Interrupt and syscall.SIGTERM are signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Warn("Received interrupt signal, shutting down...")
		cancel()
	}()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	releasesRunner := runner.NewRepoRSSPublisher(
		cfg,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	issuesRunner := runner.NewIssuesRSSPublisher(
		cfg,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	runners := []runner.Runner{
		releasesRunner, issuesRunner,
	}

	var exitErr *ExitError
	if err := executeRunners(ctx, cfg, runners); err != nil {
		if !errors.As(err, exitErr) {
			slog.Error("Error executing runners", "error", err)
		}
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Exiting...")
			return
		case <-ticker.C:
			if err := executeRunners(ctx, cfg, runners); err != nil {
				if !errors.As(err, exitErr) {
					slog.Error("Error executing runners", "error", err)
				}
				return
			}
			slog.Info("Sleeping for 24 hours...")
		}
	}
}

type ExitError struct{}

func (e ExitError) Error() string {
	return "exit"
}

func executeRunners(
	ctx context.Context,
	cfg *config.Config,
	runners []runner.Runner,
) error {
	for _, runner := range runners {
		if err := runner.Run(ctx); err != nil {
			return err
		}
	}

	if cfg.SingleRunMode {
		slog.Info("Running in single run mode, exiting...")
		return &ExitError{}
	}

	return nil
}
