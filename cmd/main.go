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

	"github.com/atomicmeganerd/starfeed/runner"
)

const (
	ghTokenKey       = "STARFEED_GITHUB_API_TOKEN"
	freshRssUrlKey   = "STARFEED_FRESHRSS_URL"
	freshRssUserKey  = "STARFEED_FRESHRSS_USER"
	freshRssTokenKey = "STARFEED_FRESHRSS_API_TOKEN"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"

	httpTimeoutInSeconds = 10
)

type Config struct {
	GithubToken   string
	FreshRssUrl   string
	FreshRssUser  string
	FreshRssToken string
	DebugMode     bool
	SingleRunMode bool
}

func NewConfig() (*Config, error) {
	// Check for required environment variables
	if os.Getenv(ghTokenKey) == "" ||
		os.Getenv(freshRssUrlKey) == "" ||
		os.Getenv(freshRssUserKey) == "" ||
		os.Getenv(freshRssTokenKey) == "" {
		slog.Error("Missing required environment variables")
		return nil, errors.New("missing required environment variables")
	}

	return &Config{
		GithubToken:   os.Getenv(ghTokenKey),
		FreshRssUrl:   os.Getenv(freshRssUrlKey),
		FreshRssUser:  os.Getenv(freshRssUserKey),
		FreshRssToken: os.Getenv(freshRssTokenKey),
		DebugMode:     os.Getenv(debugModeKey) == "true",
		SingleRunMode: os.Getenv(singleRunModeKey) == "true",
	}, nil
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("***********************************************")
	slog.Info(" Welcome to Github Releases to RSS Publisher!")
	slog.Info("***********************************************")

	cfg, err := NewConfig()
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
		cfg.GithubToken,
		cfg.FreshRssUrl,
		cfg.FreshRssUser,
		cfg.FreshRssToken,
		ctx,
		&http.Client{Timeout: httpTimeoutInSeconds * time.Second},
	)

	// Initial publish
	publisher.QueryAndPublishFeeds()
	slog.Info("Sleeping for 24 hours...")

	if cfg.SingleRunMode {
		slog.Info("Running in single run mode, exiting...")
		cancel()
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Exiting...")
			return
		case <-ticker.C:
			publisher.QueryAndPublishFeeds()
			slog.Info("Sleeping for 24 hours...")
		}
	}
}
