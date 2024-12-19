package main

import (
	"net/http"
	"os"

	"github.com/atomicmeganerd/gh-rhel-to-rss/runner"
	"github.com/charmbracelet/log"
)

const (
	ghTokenKey       = "GITHUB_API_TOKEN"
	freshRssUrlKey   = "FRESHRSS_URL"
	freshRssUserKey  = "FRESHRSS_USER"
	freshRssTokenKey = "FRESHRSS_API_TOKEN"
)

func checkForMissingEnvVar(key, value string, logger *log.Logger) {
	if value == "" {
		logger.Fatalf("Cannot run the app without the %s env var being set", key)
	}
}

func main() {

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	logger.Info("***********************************************")
	logger.Info(" Welcome to Github Releases to RSS Publisher!")
	logger.Info("***********************************************")

	debug := os.Getenv("GH_REL_TO_RSS_DEBUG")
	if debug == "true" {
		logger.Info("Debug mode enabled")
		logger.SetLevel(log.DebugLevel)
	}

	ghToken := os.Getenv(ghTokenKey)
	checkForMissingEnvVar(ghTokenKey, ghToken, logger)

	freshRssUrl := os.Getenv(freshRssUrlKey)
	checkForMissingEnvVar(freshRssUrlKey, freshRssUrl, logger)

	freshRssUser := os.Getenv(freshRssUserKey)
	checkForMissingEnvVar(freshRssUserKey, freshRssUser, logger)

	freshRssToken := os.Getenv(freshRssTokenKey)
	checkForMissingEnvVar(freshRssTokenKey, freshRssToken, logger)

	publisher := runner.NewRepoRSSPublisher(
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		http.DefaultClient,
		logger,
	)
	publisher.QueryAndPublishFeeds()

}
