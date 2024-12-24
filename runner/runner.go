package runner

import (
	"context"
	"net/http"
	"sync"
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

	log.Info("Starting main workflow....")
	start := time.Now()

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.client)

	fr := freshrss.NewFreshRSSSubManager(
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

	// Get existing subscriptions
	log.Infof("Querying existing RSS feeds in FreshRSS... ")
	rssFeeds, err := fr.GetExistingFeeds()
	if err != nil {
		log.Fatalf("Error getting list of existing feeds from FreshRSS: %s", err)
	}
	duration := time.Since(start)
	log.Infof("Queried %d feeds in FreshRSS, time: %s", len(rssFeeds), duration)

	starredRepos, err := gh.GetStarredRepos()
	if err != nil {
		log.Fatal("Could not get repos from Github: ", err)
	}
	duration = time.Since(start)
	log.Infof("Queried %d starred repos in Github, time: %s", len(starredRepos), duration)

	var wg sync.WaitGroup

	for repoFeed, repoName := range starredRepos {
		wg.Add(1)
		go p.PublishToFreshRSS(&wg, fr, at, rssFeeds, repoFeed, repoName)
	}

	for feed := range rssFeeds {
		wg.Add(1)
		go p.RemoveStaleFeeds(&wg, fr, starredRepos, feed)
	}

	wg.Wait()
	duration = time.Since(start)
	log.Infof("FreshRSS feeds synced with Github successfully, time: %s", duration)
}

func (p *RepoRSSPublisher) PublishToFreshRSS(
	wg *sync.WaitGroup,
	fr *freshrss.FreshRSSFeedManager,
	at *atom.AtomFeedChecker,
	rssFeeds map[string]struct{},
	repoFeed string,
	repoName string,
) {
	defer wg.Done()

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := rssFeeds[repoFeed]; exists {
		log.Warnf("Not adding feed %s as it is already in FreshRSS", repoFeed)
		return
	}

	if !at.CheckFeedHasEntries(repoFeed) {
		log.Warnf("Feed %s has no entries and so will not be published to RSS",
			repoFeed)
		return
	}

	err := fr.AddFeed(repoFeed, repoName, "Github")
	if err != nil {
		log.Errorf("Error publishing feed %s to FreshRSS: %s", repoFeed, err.Error())
		return
	}
}

func (p *RepoRSSPublisher) RemoveStaleFeeds(
	wg *sync.WaitGroup,
	fr *freshrss.FreshRSSFeedManager,
	repos map[string]string,
	rssFeed string,
) {
	defer wg.Done()

	// If a FreshRSS feed does not exist in Github remove it
	if _, exists := repos[rssFeed]; !exists {
		log.Infof("Removing feed %s from FreshRSS as it is no longer starred in Github", rssFeed)
		err := fr.RemoveFeed(rssFeed)

		if err != nil {
			log.Errorf("Error removing feed %s from FreshRSS: %s", rssFeed, err)
		}
	}
}
