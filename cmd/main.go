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

var logger *log.Logger

func checkForMissingEnvVar(key, value string) {
	if value == "" {
		logger.Fatalf("Cannot run the app without the %s env var being set", key)
	}
}

func main() {
	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
	})

	logger.Info("***********************************************")
	logger.Info(" Welcome to Github Releases to RSS Publisher!")
	logger.Info("***********************************************")

	ghToken := os.Getenv(ghTokenKey)
	checkForMissingEnvVar(ghTokenKey, ghToken)

	freshRssUrl := os.Getenv(freshRssUrlKey)
	checkForMissingEnvVar(freshRssUrlKey, freshRssUrl)

	freshRssUser := os.Getenv(freshRssUserKey)
	checkForMissingEnvVar(freshRssUserKey, freshRssUser)

	freshRssToken := os.Getenv(freshRssTokenKey)
	checkForMissingEnvVar(freshRssTokenKey, freshRssToken)

	publisher := runner.NewRepoRSSPublisher(ghToken, freshRssUrl, freshRssUser, freshRssToken, http.DefaultClient, logger)
	publisher.QueryAndPublishFeeds()

}
