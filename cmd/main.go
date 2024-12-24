package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicmeganerd/starfeed/runner"
)

const (
	ghTokenKey       = "GITHUB_API_TOKEN"
	freshRssUrlKey   = "FRESHRSS_URL"
	freshRssUserKey  = "FRESHRSS_USER"
	freshRssTokenKey = "FRESHRSS_API_TOKEN"
	starfeedDebugKey = "STARFEED_DEBUG"

	httpTimeoutInSeconds = 10
)

func checkForMissingEnvVar(key, value string, sigChan chan<- os.Signal) {
	if value == "" {
		slog.Error("Cannot run this app without the env var being set", "key", key)
		sigChan <- syscall.SIGTERM
	}
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.Info("***********************************************")
	slog.Info(" Welcome to Github Releases to RSS Publisher!")
	slog.Info("***********************************************")

	debug := os.Getenv(starfeedDebugKey)
	handler := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if debug == "true" {
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

	ghToken := os.Getenv(ghTokenKey)
	checkForMissingEnvVar(ghTokenKey, ghToken, sigChan)

	freshRssUrl := os.Getenv(freshRssUrlKey)
	checkForMissingEnvVar(freshRssUrlKey, freshRssUrl, sigChan)

	freshRssUser := os.Getenv(freshRssUserKey)
	checkForMissingEnvVar(freshRssUserKey, freshRssUser, sigChan)

	freshRssToken := os.Getenv(freshRssTokenKey)
	checkForMissingEnvVar(freshRssTokenKey, freshRssToken, sigChan)

	publisher := runner.NewRepoRSSPublisher(
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		ctx,
		&http.Client{Timeout: httpTimeoutInSeconds * time.Second},
	)

	// Initial publish
	publisher.QueryAndPublishFeeds()
	slog.Info("Sleeping for 24 hours...")

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
