package runner

import (
	"context"
	"net/http"
	"sync"
	"time"

	"log/slog"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/freshrss"
	"github.com/atomicmeganerd/starfeed/github"
)

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type RepoRSSPublisher struct {
	ghToken       string // WARNING: Do not log this value as it is a secret
	freshRssUrl   string
	freshRssUser  string
	freshRssToken string // WARNING: Do not logger this value as it is a secret
	ctx           context.Context
	client        *http.Client
}

// NewRepoRSSPublisher creates a new RepoRSSPublisher instance.
// Arguments:
// - ghToken: The GitHub API token to authenticate with.
// - freshRssUrl: The base URL of the FreshRSS instance.
// - freshRssUser: The username to authenticate to FreshRSS.
// - freshRssToken: The API token to authenticate with FreshRSS.
// - ctx: The context to use for requests.
// - client: The http client to use for requests (used for mocking).
func NewRepoRSSPublisher(ghToken, freshRssUrl, freshRssUser, freshRssToken string,
	ctx context.Context,
	client *http.Client) RepoRSSPublisher {
	return RepoRSSPublisher{
		ghToken,
		freshRssUrl,
		freshRssUser,
		freshRssToken,
		ctx,
		client,
	}
}

// QueryAndPublishFeeds queries the starred repos from GitHub and publishes them to FreshRSS.
// It also removes any stale feeds from FreshRSS as long as they are not starred in GitHub but
// are actually Github release feeds.
func (p *RepoRSSPublisher) QueryAndPublishFeeds() {
	slog.Info("Starting main workflow....")
	start := time.Now()

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.ctx, p.client)
	fr := freshrss.NewFreshRSSFeedManager(
		p.freshRssUrl, p.freshRssUser, p.freshRssToken, p.ctx, p.client,
	)
	at := atom.NewAtomFeedChecker(p.ctx, p.client)

	// Authenticate to FreshRSS
	if err := fr.Authenticate(); err != nil {
		slog.Error("Could not authenticate with FreshRSS", "error", err.Error())
		return
	}

	// Get existing subscriptions
	slog.Info("Querying existing RSS feeds in FreshRSS... ")
	rssFeedMap, err := fr.GetExistingFeeds()
	if err != nil {
		slog.Error("Error getting list of existing feeds from FreshRSS", "error", err.Error())
		return
	}
	// Filter out any subscriptions that are not Github release feeds so we
	// do not unsubscribe from them
	rssFeedMap = filterOutNonGithubFeeds(gh, rssFeedMap)
	duration := time.Since(start)
	slog.Info(
		"Queried Github release feeds in FreshRSS",
		"numberFeeds",
		len(rssFeedMap),
		"duration",
		duration,
	)

	// Get starred repos from Github
	starredRepoMap, err := gh.GetStarredRepos()
	if err != nil {
		slog.Error("Could not get repos from Github", "error", err.Error())
		return
	}
	duration = time.Since(start)
	slog.Info(
		"Queried starred repos in Github", "numberStarredRepos", len(starredRepoMap),
		"duration", duration,
	)

	// Sync feeds
	var wg sync.WaitGroup
	for _, repo := range starredRepoMap {
		wg.Add(1)
		go publishToFreshRSS(&wg, fr, at, rssFeedMap, repo)
	}
	for feed := range rssFeedMap {
		wg.Add(1)
		go removeStaleFeeds(&wg, fr, starredRepoMap, feed)
	}
	wg.Wait()

	// Report success
	duration = time.Since(start)
	slog.Info("FreshRSS feeds synced with Github successfully", "duration", duration)
}

func publishToFreshRSS(
	wg *sync.WaitGroup,
	fr freshrss.FreshRSSFeedManager,
	at atom.AtomFeedChecker,
	rssFeedMap map[string]struct{},
	repo github.GitHubRepo,
) {
	defer wg.Done()

	repoFeed := repo.FeedUrl

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := rssFeedMap[repoFeed]; exists {
		slog.Info("Not adding feed as it is already in FreshRSS", "feed", repoFeed)
		return
	}

	if !at.CheckFeedHasEntries(repoFeed) {
		slog.Info("Not adding feed as it has zero entries", "feed", repoFeed)
		return
	}

	if err := fr.AddFeed(repoFeed, repo.Name, "Github"); err != nil {
		slog.Error("Error publishing feed to FreshRSS", "feed", repoFeed, "error", err.Error())
		return
	}
}

func filterOutNonGithubFeeds(
	gh *github.GitHubStarredFeedBuilder,
	rssFeedMap map[string]struct{},
) map[string]struct{} {
	filteredMap := make(map[string]struct{})
	for k, v := range rssFeedMap {
		if gh.IsGithubReleasesFeed(k) {
			filteredMap[k] = v
		} else {
			slog.Debug("Removing non-Github feed from RSS map so we don't unsubscribe", "feed", k)
		}
	}
	return filteredMap
}

func removeStaleFeeds(
	wg *sync.WaitGroup,
	fr freshrss.FreshRSSFeedManager,
	starredRepoMap map[string]github.GitHubRepo, // The key is the release ATOM feed
	rssFeed string,
) {
	defer wg.Done()

	// If a FreshRSS feed does not exist in Github remove it
	if _, exists := starredRepoMap[rssFeed]; !exists {
		slog.Info(
			"Removing feed from FreshRSS as it is no longer starred in Github", "feed", rssFeed,
		)
		if err := fr.RemoveFeed(rssFeed); err != nil {
			slog.Error("Error removing feed from FreshRSS", "feed", rssFeed, "Error", err.Error())
		}
	}
}
