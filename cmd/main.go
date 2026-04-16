package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/runner"
)

const (
	DISABLE_REPO = true
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

	publisher := runner.NewRepoRSSPublisher(
		cfg.GitHubToken,
		cfg.FreshRSSURL,
		cfg.FreshRSSUser,
		cfg.FreshRSSToken,
		ctx,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	issuesPublisher := runner.NewIssuesRSSPublisher(
		cfg.GitHubToken,
		ctx,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	if !DISABLE_REPO {
		if err := publisher.QueryAndPublishFeeds(); err != nil {
			slog.Error("Error with repo feeds workflow", "error", err)
		}
	}

	if err := issuesPublisher.QueryAndPublishFeeds(); err != nil {
		slog.Error("Error with issues feed workflow", "error", err)
	}

	if cfg.SingleRunMode {
		slog.Info("Running in single run mode, exiting...")
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Exiting...")
			return
		case <-ticker.C:
			if err := publisher.QueryAndPublishFeeds(); err != nil {
				slog.Error("Error with repo feeds workflow", "error", err)
			}
			slog.Info("Sleeping for 24 hours...")
		}
	}
}
