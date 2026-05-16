package runner

import (
	"context"
	"log/slog"
	"time"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/rss"
	"golang.org/x/sync/errgroup"
)

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type publishReleasesRunner struct {
	gitHost         githost.GitHost
	rssServer       rss.RSSServer
	atomFeedChecker atom.AtomFeedChecker
}

// NewPublishReleasesRunner creates a new RepoRSSPublisher instance.
// Arguments:
// - cfg: the config object that holds all of the relevant configuration.
// - client: The http client to use for requests (used for mocking).
func NewPublishReleasesRunner(
	gitHost githost.GitHost,
	rssServer rss.RSSServer,
	atomFeedChecker atom.AtomFeedChecker,
) Runner {
	return &publishReleasesRunner{
		gitHost,
		rssServer,
		atomFeedChecker,
	}
}

// Run queries the starred repos from GitHub and publishes them to FreshRSS.
// It also removes any stale feeds from FreshRSS as long as they are not starred in GitHub but
// are actually GitHub release feeds.
func (p *publishReleasesRunner) Run(ctx context.Context) error {
	slog.Info("Starting main workflow...")
	start := time.Now()

	// Authenticate to FreshRSS
	if err := p.rssServer.Authenticate(ctx); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	var rssFeedMap map[string]rss.RSSFeed
	var starredRepoMap map[string]githost.Repo

	// Get existing subscriptions
	g.Go(func() error {
		slog.Info("Querying existing RSS feeds in FreshRSS... ")
		feeds, err := p.rssServer.GetExistingFeeds(ctx)
		if err != nil {
			return err
		}

		// Filter out any subscriptions that are not GitHub release feeds so we
		// do not unsubscribe from them
		rssFeedMap = filterOutNonGitHubFeeds(p.gitHost, feeds)
		slog.Info(
			"Queried GitHub release feeds in FreshRSS",
			"numberFeeds", len(rssFeedMap),
			"duration", time.Since(start),
		)
		return nil
	})

	// Get starred repos from the Git provider
	g.Go(func() error {
		repos, err := p.gitHost.GetStarredRepos(ctx)
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
			return p.publishToFreshRSS(ctx, rssFeedMap, repo)
		})
	}
	for feed := range rssFeedMap {
		g.Go(func() error {
			return p.removeStaleFeeds(ctx, starredRepoMap, feed)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Report success
	slog.Info("FreshRSS feeds synced with GitHub successfully", "duration", time.Since(start))
	return nil
}

func (p *publishReleasesRunner) publishToFreshRSS(
	ctx context.Context,
	rssFeedMap map[string]rss.RSSFeed,
	repo githost.Repo,
) error {
	repoFeed := repo.FeedURL()

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := rssFeedMap[repoFeed]; exists {
		slog.Info("Not adding feed as it is already in FreshRSS", "feed", repo.Name())
		return nil
	}

	hasEntries, err := p.atomFeedChecker.CheckFeedHasEntries(ctx, repoFeed)
	if err != nil {
		return err
	}

	if !hasEntries {
		slog.Info("Not adding feed as it has zero entries", "feed", repo.Name())
		return nil
	}

	return p.rssServer.AddFeed(ctx, repoFeed, repo.Name(), p.gitHost.Name())
}

// We never want to unsubscribe from non-github feeds.
func filterOutNonGitHubFeeds(
	gh githost.GitHost,
	rssFeedMap map[string]rss.RSSFeed,
) map[string]rss.RSSFeed {
	filteredMap := make(map[string]rss.RSSFeed)
	for k, v := range rssFeedMap {
		if gh.IsReleaseFeed(k) {
			filteredMap[k] = v
		} else {
			slog.Debug("Ignoring non-GitHub feed so we don't unsubscribe", "feed", k)
		}
	}
	return filteredMap
}

func (p *publishReleasesRunner) removeStaleFeeds(
	ctx context.Context,
	starredRepoMap map[string]githost.Repo, // The key is the release ATOM feed
	rssFeed string,
) error {
	// If a FreshRSS feed does not exist in GitHub remove it
	if _, exists := starredRepoMap[rssFeed]; !exists {
		slog.Info(
			"Removing feed from FreshRSS as it is no longer starred in GitHub", "feed", rssFeed,
		)
		if err := p.rssServer.RemoveFeed(ctx, rssFeed); err != nil {
			return err
		}
	}
	return nil
}
