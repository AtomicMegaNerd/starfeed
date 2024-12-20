package main

import (
	"net/http"
	"os"

	"github.com/atomicmeganerd/starfeed/runner"
	"github.com/charmbracelet/log"
)

const (
	ghTokenKey       = "GITHUB_API_TOKEN"
	freshRssUrlKey   = "FRESHRSS_URL"
	freshRssUserKey  = "FRESHRSS_USER"
	freshRssTokenKey = "FRESHRSS_API_TOKEN"
)

func checkForMissingEnvVar(key, value string) {
	if value == "" {
		log.Fatalf("Cannot run the app without the %s env var being set", key)
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

	log.Info("***********************************************")
	log.Info(" Welcome to Github Releases to RSS Publisher!")
	log.Info("***********************************************")

	debug := os.Getenv("STARFEED_DEBUG")
	if debug == "true" {
		log.Info("Debug mode enabled")
		log.SetLevel(log.DebugLevel)
	}

	ghToken := os.Getenv(ghTokenKey)
	checkForMissingEnvVar(ghTokenKey, ghToken)

	freshRssUrl := os.Getenv(freshRssUrlKey)
	checkForMissingEnvVar(freshRssUrlKey, freshRssUrl)

	freshRssUser := os.Getenv(freshRssUserKey)
	checkForMissingEnvVar(freshRssUserKey, freshRssUser)

	freshRssToken := os.Getenv(freshRssTokenKey)
	checkForMissingEnvVar(freshRssTokenKey, freshRssToken)

	publisher := runner.NewRepoRSSPublisher(
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		http.DefaultClient,
	)
	publisher.QueryAndPublishFeeds()

}
