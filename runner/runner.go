package runner

import (
	"net/http"

	"github.com/atomicmeganerd/gh-rhel-to-rss/github"
	"github.com/charmbracelet/log"
)

type RepoRSSPublisher struct {
	ghToken       string // WARNING: Do not log this value as it is a secret
	freshRssUrl   string
	freshRssUser  string
	freshRssToken string // WARNING: Do not log this value as it is a secret
	client        *http.Client
	logger        *log.Logger
}

func NewRepoRSSPublisher(ghToken, freshRssUrl, freshRssUser, freshRssToken string,
	client *http.Client, logger *log.Logger) RepoRSSPublisher {
	return RepoRSSPublisher{
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		client,
		logger,
	}
}

func (p *RepoRSSPublisher) QueryAndPublishFeeds() {

	logger := p.logger

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, http.DefaultClient, logger)

	repos, err := gh.GetStarredRepos()

	if err != nil {
		log.Fatal("Could not get repos from Github: ", err)
	}

	for _, repo := range repos {
		log.Infof("Found starred repo: %s", repo.String())
	}

}
