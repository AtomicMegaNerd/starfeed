package runner

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/freshrss"
	"github.com/atomicmeganerd/starfeed/github"
)

type RepoRSSPublisher interface {
	QueryAndPublishFeeds()
}

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type repoRSSPublisher struct {
	ghToken       string // WARNING: Do not log this value as it is a secret
	freshRSSURL   string
	freshRSSUser  string
	freshRSSToken string // WARNING: Do not log this value as it is a secret
	ctx           context.Context
	client        *http.Client
}

// NewRepoRSSPublisher creates a new RepoRSSPublisher instance.
// Arguments:
// - ghToken: The GitHub API token to authenticate with.
// - freshRSSUrl: The base URL of the FreshRSS instance.
// - freshRSSUser: The username to authenticate to FreshRSS.
// - freshRSSToken: The API token to authenticate with FreshRSS.
// - ctx: The context to use for requests.
// - client: The http client to use for requests (used for mocking).
func NewRepoRSSPublisher(ghToken, freshRSSURL, freshRSSUser, freshRSSToken string,
	ctx context.Context,
	client *http.Client,
) RepoRSSPublisher {
	return &repoRSSPublisher{
		ghToken,
		freshRSSURL,
		freshRSSUser,
		freshRSSToken,
		ctx,
		client,
	}
}

// QueryAndPublishFeeds queries the starred repos from GitHub and publishes them to FreshRSS.
// It also removes any stale feeds from FreshRSS as long as they are not starred in GitHub but
// are actually GitHub release feeds.
func (p *repoRSSPublisher) QueryAndPublishFeeds() {
	slog.Info("Starting main workflow....")
	start := time.Now()

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.ctx, p.client)
	fr := freshrss.NewFreshRSSFeedManager(
		p.freshRSSURL, p.freshRSSUser, p.freshRSSToken, p.ctx, p.client,
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
	// Filter out any subscriptions that are not GitHub release feeds so we
	// do not unsubscribe from them
	rssFeedMap = filterOutNonGitHubFeeds(gh, rssFeedMap)
	duration := time.Since(start)
	slog.Info(
		"Queried GitHub release feeds in FreshRSS",
		"numberFeeds",
		len(rssFeedMap),
		"duration",
		duration,
	)

	// Get starred repos from GitHub
	starredRepoMap, err := gh.GetStarredRepos()
	if err != nil {
		slog.Error("Error getting repos from GitHub", "error", err.Error())
		return
	}
	duration = time.Since(start)
	slog.Info(
		"Queried starred repos in GitHub", "numberStarredRepos", len(starredRepoMap),
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
	slog.Info("FreshRSS feeds synced with GitHub successfully", "duration", duration)
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

	hasEntries, err := at.CheckFeedHasEntries(repoFeed)
	if err != nil {
		slog.Error("Error checking if feed has entries", "feed", repoFeed, "error", err.Error())
		return
	}

	if !hasEntries {
		slog.Info("Not adding feed as it has zero entries", "feed", repoFeed)
		return
	}

	if err := fr.AddFeed(repoFeed, repo.Name, "GitHub"); err != nil {
		slog.Error("Error publishing feed to FreshRSS", "feed", repoFeed, "error", err.Error())
		return
	}
}

func filterOutNonGitHubFeeds(
	gh github.GitHubStarredFeedBuilder,
	rssFeedMap map[string]struct{},
) map[string]struct{} {
	filteredMap := make(map[string]struct{})
	for k, v := range rssFeedMap {
		if gh.IsGitHubReleasesFeed(k) {
			filteredMap[k] = v
		} else {
			slog.Debug("Removing non-GitHub feed from RSS map so we don't unsubscribe", "feed", k)
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

	// If a FreshRSS feed does not exist in GitHub remove it
	if _, exists := starredRepoMap[rssFeed]; !exists {
		slog.Info(
			"Removing feed from FreshRSS as it is no longer starred in GitHub", "feed", rssFeed,
		)
		if err := fr.RemoveFeed(rssFeed); err != nil {
			slog.Error("Error removing feed from FreshRSS", "feed", rssFeed, "error", err.Error())
		}
	}
}
