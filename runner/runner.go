package runner

import (
	"net/http"
	"sync"

	"github.com/atomicmeganerd/gh-rhel-to-rss/freshrss"
	"github.com/atomicmeganerd/gh-rhel-to-rss/github"
	"github.com/charmbracelet/log"
)

type RepoRSSPublisher struct {
	ghToken       string // WARNING: Do not logger.this value as it is a secret
	freshRssUrl   string
	freshRssUser  string
	freshRssToken string // WARNING: Do not logger.this value as it is a secret
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

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.client, p.logger)

	fr := freshrss.NewFreshRSSFeedPublisher(
		p.freshRssUrl,
		p.freshRssUser,
		p.freshRssToken,
		p.client,
		p.logger,
	)

	err := fr.Authenticate()
	if err != nil {
		logger.Fatalf("Could not authenticate with FreshRSS: %s", err)
	}

	starredRepos, err := gh.GetStarredRepos()

	if err != nil {
		logger.Fatal("Could not get repos from Github: ", err)
	}

	var wg sync.WaitGroup
	wg.Add(len(starredRepos))

	for _, repo := range starredRepos {
		logger.Infof("Found starred repo: %s", repo.String())
		go p.AddToFreshRSS(&wg, fr, &repo)
	}

	wg.Wait()
	logger.Info("All feeds published to FreshRSS")
}

func (p *RepoRSSPublisher) AddToFreshRSS(
	wg *sync.WaitGroup,
	fr *freshrss.FreshRSSFeedPublisher,
	repo *github.GitHubRepo,
) {
	defer wg.Done()

	err := fr.AddFeed(repo.ReleasesAtomFeed, repo.Name, "Github")
	if err != nil {
		p.logger.Errorf("Error publishing feed %s to FreshRSS: %s", repo.ReleasesAtomFeed, err.Error())
		return
	}
}
