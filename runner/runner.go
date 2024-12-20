package runner

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/sync/semaphore"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/freshrss"
	"github.com/atomicmeganerd/starfeed/github"
	"github.com/charmbracelet/log"
)

type RepoRSSPublisher struct {
	ghToken       string // WARNING: Do not logger.this value as it is a secret
	freshRssUrl   string
	freshRssUser  string
	freshRssToken string // WARNING: Do not logger.this value as it is a secret
	ctx           context.Context
	client        *http.Client
}

func NewRepoRSSPublisher(ghToken, freshRssUrl, freshRssUser, freshRssToken string,
	client *http.Client) RepoRSSPublisher {
	ctx := context.Background()
	return RepoRSSPublisher{
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		ctx,
		client,
	}
}

func (p *RepoRSSPublisher) QueryAndPublishFeeds() {

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.client)

	fr := freshrss.NewFreshRSSFeedPublisher(
		p.freshRssUrl,
		p.freshRssUser,
		p.freshRssToken,
		p.client,
	)

	at := atom.NewAtomFeedChecker(p.client)

	err := fr.Authenticate()
	if err != nil {
		log.Fatalf("Could not authenticate with FreshRSS: %s", err)
	}

	starredRepos, err := gh.GetStarredRepos()

	if err != nil {
		log.Fatal("Could not get repos from Github: ", err)
	}

	var wg sync.WaitGroup

	// Limit the number of concurrent requests to FreshRSS so we don't smack it too hard
	sem := semaphore.NewWeighted(int64(3))

	for _, repo := range starredRepos {
		if at.CheckFeedHasEntries(repo.ReleasesAtomFeed) {
			log.Infof(
				"Found starred repo %s that has a valid release feed. Publishing to FreshRSS...",
				repo.Name,
			)
			wg.Add(1)
			err := sem.Acquire(p.ctx, 1)
			if err != nil {
				log.Errorf("Error acquiring semaphore: %s", err)
				return
			}
			go p.AddToFreshRSS(&wg, sem, fr, &repo)
		}
	}

	wg.Wait()
	log.Info("All feeds published to FreshRSS")
}

func (p *RepoRSSPublisher) AddToFreshRSS(
	wg *sync.WaitGroup,
	sem *semaphore.Weighted,
	fr *freshrss.FreshRSSFeedPublisher,
	repo *github.GitHubRepo,
) {
	defer wg.Done()
	defer sem.Release(1)

	err := fr.AddFeed(repo.ReleasesAtomFeed, repo.Name, "Github")
	if err != nil {
		log.Errorf("Error publishing feed %s to FreshRSS: %s", repo.ReleasesAtomFeed, err.Error())
		return
	}
}
