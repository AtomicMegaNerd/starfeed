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
	log.Infof("Querying existing subs in FreshRSS... ")
	subMap, err := fr.GetExistingFeeds()
	if err != nil {
		log.Fatalf("Error getting list of existing subs from FreshRSS: %s", err)
	}
	duration := time.Since(start)
	log.Infof("Queried existing subs in FreshRSS, time: %s", duration)

	starredRepos, err := gh.GetStarredRepos()
	if err != nil {
		log.Fatal("Could not get repos from Github: ", err)
	}
	duration = time.Since(start)
	log.Infof("Queried list of starred repos in Github, time: %s", duration)

	var wg sync.WaitGroup
	for _, repo := range starredRepos {
		wg.Add(1)
		go p.PublishToFreshRSS(&wg, fr, at, subMap, &repo)
	}

	wg.Wait()
	duration = time.Since(start)
	log.Infof("All feeds published to FreshRSS, time: %s", duration)
}

func (p *RepoRSSPublisher) PublishToFreshRSS(
	wg *sync.WaitGroup,
	fr *freshrss.FreshRSSFeedManager,
	at *atom.AtomFeedChecker,
	subMap map[string]struct{},
	repo *github.GitHubRepo,
) {
	defer wg.Done()
	feedUrl := repo.ReleasesFeedUrl

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := subMap[feedUrl]; exists {
		log.Warnf("Not adding feed %s as it is already in FreshRSS", feedUrl)
		return
	}

	if !at.CheckFeedHasEntries(feedUrl) {
		log.Warnf("Feed %s has no entries and so will not be published to RSS",
			repo.ReleasesFeedUrl)
		return
	}

	err := fr.AddFeed(feedUrl, repo.Name, "Github")
	if err != nil {
		log.Errorf("Error publishing feed %s to FreshRSS: %s", repo.ReleasesFeedUrl, err.Error())
		return
	}
}
