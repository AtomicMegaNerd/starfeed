package runner

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/freshrss"
	"github.com/atomicmeganerd/starfeed/githost"
	"golang.org/x/sync/errgroup"
)

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type publishReleasesRunner struct {
	gitHost githost.GitHostConfig
	cfg     *config.Config
	client  *http.Client
}

// NewPublishReleasesRunner creates a new RepoRSSPublisher instance.
// Arguments:
// - cfg: the config object that holds all of the relevant configuration.
// - client: The http client to use for requests (used for mocking).
func NewPublishReleasesRunner(
	gitHost githost.GitHostConfig,
	cfg *config.Config,
	client *http.Client,
) Runner {
	return &publishReleasesRunner{
		gitHost,
		cfg,
		client,
	}
}

// Run queries the starred repos from GitHub and publishes them to FreshRSS.
// It also removes any stale feeds from FreshRSS as long as they are not starred in GitHub but
// are actually GitHub release feeds.
func (p *publishReleasesRunner) Run(ctx context.Context) error {
	slog.Info("Starting main workflow...")
	start := time.Now()

	gh, err := GetGitHostFromConfig(p.gitHost, p.client)
	if err != nil {
		return err
	}
	fr := freshrss.NewFreshRSSFeedManager(p.cfg, p.client)
	at := atom.NewAtomFeedChecker(p.client)

	// Authenticate to FreshRSS
	if err := fr.Authenticate(ctx); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	var rssFeedMap map[string]freshrss.RSSFeed
	var starredRepoMap map[string]githost.Repo

	// Get existing subscriptions
	g.Go(func() error {
		slog.Info("Querying existing RSS feeds in FreshRSS... ")
		feeds, err := fr.GetExistingFeeds(ctx)
		if err != nil {
			return err
		}

		// Filter out any subscriptions that are not GitHub release feeds so we
		// do not unsubscribe from them
		rssFeedMap = filterOutNonGitHubFeeds(gh, feeds)
		slog.Info(
			"Queried GitHub release feeds in FreshRSS",
			"numberFeeds", len(rssFeedMap),
			"duration", time.Since(start),
		)
		return nil
	})

	// Get starred repos from the Git provider
	g.Go(func() error {
		repos, err := gh.GetStarredRepos(ctx)
		if err != nil {
			return err
		}
		starredRepoMap = repos

		slog.Info(
			"Queried starred repos in GitHub",
			"numberStarredRepos", len(starredRepoMap),
			"duration", time.Since(start),
		)
		return nil
	})

	// Wait for these two independent operations to finish...
	if err := g.Wait(); err != nil {
		return err
	}

	// Sync feeds using an error group with concurrency limit to avoid overwhelming FreshRSS
	g, ctx = errgroup.WithContext(ctx)
	g.SetLimit(5)

	for _, repo := range starredRepoMap {
		g.Go(func() error {
			return publishToFreshRSS(ctx, fr, at, rssFeedMap, repo, string(p.gitHost.Type))
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
	slog.Info("FreshRSS feeds synced with GitHub successfully", "duration", time.Since(start))
	return nil
}

func publishToFreshRSS(
	ctx context.Context,
	fr freshrss.FreshRSSFeedManager,
	at atom.AtomFeedChecker,
	rssFeedMap map[string]freshrss.RSSFeed,
	repo githost.Repo,
	feedName string,
) error {
	repoFeed := repo.FeedURL()

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

	return fr.AddFeed(ctx, repoFeed, repo.Name(), feedName)
}

// We never want to unsubscribe from non-github feeds.
func filterOutNonGitHubFeeds(
	gh githost.GitHost,
	rssFeedMap map[string]freshrss.RSSFeed,
) map[string]freshrss.RSSFeed {
	filteredMap := make(map[string]freshrss.RSSFeed)
	for k, v := range rssFeedMap {
		if gh.IsReleaseFeed(k) {
			filteredMap[k] = v
		} else {
			slog.Debug("Ignoring non-GitHub feed so we don't unsubscribe", "feed", k)
		}
	}
	return filteredMap
}

func removeStaleFeeds(
	ctx context.Context,
	fr freshrss.FreshRSSFeedManager,
	starredRepoMap map[string]githost.Repo, // The key is the release ATOM feed
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
