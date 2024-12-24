package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicmeganerd/starfeed/runner"
	"github.com/charmbracelet/log"
)

const (
	ghTokenKey       = "GITHUB_API_TOKEN"
	freshRssUrlKey   = "FRESHRSS_URL"
	freshRssUserKey  = "FRESHRSS_USER"
	freshRssTokenKey = "FRESHRSS_API_TOKEN"
	starfeedDebugKey = "STARFEED_DEBUG"
)

func checkForMissingEnvVar(key, value string, sigChan chan<- os.Signal) {
	if value == "" {
		log.Error("Cannot run this app without the %s env var being set", key)
		sigChan <- syscall.SIGTERM
	}
}

func init() {
	log.SetDefault(
		log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
		}),
	)
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("***********************************************")
	log.Info(" Welcome to Github Releases to RSS Publisher!")
	log.Info("***********************************************")

	debug := os.Getenv(starfeedDebugKey)
	if debug == "true" {
		log.Info("Debug mode enabled")
		log.SetLevel(log.DebugLevel)
	}

	// In this case both os.Interrupt and syscall.SIGTERM are signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Warn("Received interrupt signal, shutting down...")
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
		sigChan,
		http.DefaultClient,
	)

	// Initial publish
	publisher.QueryAndPublishFeeds()
	log.Info("Sleeping for 24 hours...")

	for {
		select {
		case <-ctx.Done():
			log.Info("Exiting...")
			return
		case <-ticker.C:
			publisher.QueryAndPublishFeeds()
			log.Info("Sleeping for 24 hours...")
		}
	}
}
