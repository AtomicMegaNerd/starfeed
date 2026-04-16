package runner

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/freshrss"
	"github.com/atomicmeganerd/starfeed/github"
	"golang.org/x/sync/errgroup"
)

type RepoRSSPublisher interface {
	QueryAndPublishFeeds(ctx context.Context) error
}

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type repoRSSPublisher struct {
	ghToken       string // WARNING: Do not log this value as it is a secret
	freshRSSURL   string
	freshRSSUser  string
	freshRSSToken string // WARNING: Do not log this value as it is a secret
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
	client *http.Client,
) RepoRSSPublisher {
	return &repoRSSPublisher{
		ghToken,
		freshRSSURL,
		freshRSSUser,
		freshRSSToken,
		client,
	}
}

// QueryAndPublishFeeds queries the starred repos from GitHub and publishes them to FreshRSS.
// It also removes any stale feeds from FreshRSS as long as they are not starred in GitHub but
// are actually GitHub release feeds.
func (p *repoRSSPublisher) QueryAndPublishFeeds(ctx context.Context) error {
	slog.Info("Starting main workflow....")
	start := time.Now()

	gh := github.NewGitHubStarredFeedBuilder(p.ghToken, p.client)
	fr := freshrss.NewFreshRSSFeedManager(
		p.freshRSSURL, p.freshRSSUser, p.freshRSSToken, p.client,
	)
	at := atom.NewAtomFeedChecker(p.client)

	// Authenticate to FreshRSS
	if err := fr.Authenticate(ctx); err != nil {
		return err
	}

	// Get existing subscriptions
	slog.Info("Querying existing RSS feeds in FreshRSS... ")
	rssFeedMap, err := fr.GetExistingFeeds(ctx)
	if err != nil {
		return err
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
	starredRepoMap, err := gh.GetStarredRepos(ctx)
	if err != nil {
		return err
	}
	duration = time.Since(start)
	slog.Info(
		"Queried starred repos in GitHub", "numberStarredRepos", len(starredRepoMap),
		"duration", duration,
	)

	// Sync feeds using an error group
	g, ctx := errgroup.WithContext(ctx)
	for _, repo := range starredRepoMap {
		g.Go(func() error {
			return publishToFreshRSS(ctx, fr, at, rssFeedMap, repo)
		})
	}
	for feed := range rssFeedMap {
		g.Go(func() error {
			return removeStaleFeeds(ctx, fr, starredRepoMap, feed)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Report success
	duration = time.Since(start)
	slog.Info("FreshRSS feeds synced with GitHub successfully", "duration", duration)
	return nil
}

func publishToFreshRSS(
	ctx context.Context,
	fr freshrss.FreshRSSFeedManager,
	at atom.AtomFeedChecker,
	rssFeedMap map[string]struct{},
	repo github.GitHubRepo,
) error {
	repoFeed := repo.ReleaseFeedUrl

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := rssFeedMap[repoFeed]; exists {
		slog.Info("Not adding feed as it is already in FreshRSS", "feed", repoFeed)
		return nil
	}

	hasEntries, err := at.CheckFeedHasEntries(ctx, repoFeed)
	if err != nil {
		return err
	}

	if !hasEntries {
		slog.Info("Not adding feed as it has zero entries", "feed", repoFeed)
		return nil
	}

	if err := fr.AddFeed(ctx, repoFeed, repo.Name, "GitHub"); err != nil {
		return err
	}
	return nil
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
	ctx context.Context,
	fr freshrss.FreshRSSFeedManager,
	starredRepoMap map[string]github.GitHubRepo, // The key is the release ATOM feed
	rssFeed string,
) error {
	// If a FreshRSS feed does not exist in GitHub remove it
	if _, exists := starredRepoMap[rssFeed]; !exists {
		slog.Info(
			"Removing feed from FreshRSS as it is no longer starred in GitHub", "feed", rssFeed,
		)
		if err := fr.RemoveFeed(ctx, rssFeed); err != nil {
			return err
		}
	}
	return nil
}
