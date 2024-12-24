package runner

import (
	"context"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

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
	sigChan       chan<- os.Signal
	client        *http.Client
}

func NewRepoRSSPublisher(ghToken, freshRssUrl, freshRssUser, freshRssToken string,
	ctx context.Context,
	sigChan chan<- os.Signal,
	client *http.Client) RepoRSSPublisher {
	return RepoRSSPublisher{
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		ctx,
		sigChan,
		client,
	}
}

func (p *RepoRSSPublisher) QueryAndPublishFeeds() {
	log.Info("Starting main workflow....")
	start := time.Now()

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.ctx, p.client)
	fr := freshrss.NewFreshRSSFeedManager(
		p.freshRssUrl, p.freshRssUser, p.freshRssToken, p.ctx, p.client,
	)
	at := atom.NewAtomFeedChecker(p.ctx, p.client)

	if err := fr.Authenticate(); err != nil {
		log.Errorf("Could not authenticate with FreshRSS: %s", err)
		p.sigChan <- syscall.SIGTERM
		return
	}

	// Get existing subscriptions
	log.Infof("Querying existing RSS feeds in FreshRSS... ")
	rssFeedMap, err := fr.GetExistingFeeds()
	if err != nil {
		log.Errorf("Error getting list of existing feeds from FreshRSS: %s", err)
		p.sigChan <- syscall.SIGTERM
		return
	}

	duration := time.Since(start)
	log.Infof("Queried %d feeds in FreshRSS, time: %s", len(rssFeedMap), duration)

	// Get starred repos
	starredRepoMap, err := gh.GetStarredRepos()
	if err != nil {
		log.Errorf("Could not get repos from Github: %s", err)
		p.sigChan <- syscall.SIGTERM
		return
	}
	duration = time.Since(start)
	log.Infof("Queried %d starred repos in Github, time: %s", len(starredRepoMap), duration)

	// Sync feeds
	var wg sync.WaitGroup
	for _, repo := range starredRepoMap {
		wg.Add(1)
		go p.PublishToFreshRSS(&wg, fr, at, rssFeedMap, repo)
	}
	for feed := range rssFeedMap {
		wg.Add(1)
		go p.RemoveStaleFeeds(&wg, fr, starredRepoMap, feed)
	}
	wg.Wait()

	// Report success
	duration = time.Since(start)
	log.Infof("FreshRSS feeds synced with Github successfully, time: %s", duration)
}

func (p *RepoRSSPublisher) PublishToFreshRSS(
	wg *sync.WaitGroup,
	fr *freshrss.FreshRSSFeedManager,
	at *atom.AtomFeedChecker,
	rssFeedMap map[string]struct{},
	repo github.GitHubRepo,
) {
	defer wg.Done()

	repoFeed := repo.FeedUrl

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := rssFeedMap[repoFeed]; exists {
		log.Warnf("Not adding feed %s as it is already in FreshRSS", repoFeed)
		return
	}

	if !at.CheckFeedHasEntries(repoFeed) {
		log.Warnf("Feed %s has no entries and so will not be published to RSS", repoFeed)
		return
	}

	if err := fr.AddFeed(repoFeed, repo.Name, "Github"); err != nil {
		log.Errorf("Error publishing feed %s to FreshRSS: %s", repoFeed, err.Error())
		return
	}
}

func (p *RepoRSSPublisher) RemoveStaleFeeds(
	wg *sync.WaitGroup,
	fr *freshrss.FreshRSSFeedManager,
	starredRepoMap map[string]github.GitHubRepo, // The key is the release ATOM feed
	rssFeed string,
) {
	defer wg.Done()

	// If a FreshRSS feed does not exist in Github remove it
	if _, exists := starredRepoMap[rssFeed]; !exists {
		log.Infof("Removing feed %s from FreshRSS as it is no longer starred in Github", rssFeed)
		if err := fr.RemoveFeed(rssFeed); err != nil {
			log.Errorf("Error removing feed %s from FreshRSS: %s", rssFeed, err)
		}
	}
}
