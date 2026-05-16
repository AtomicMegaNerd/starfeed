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

// This queries release feeds for all starred repos in the specified Git host and publishes them
// to FreshRSS. It also removes any stale release feeds from FreshRSS if they are no longer
// starred.
func (p *publishReleasesRunner) Run(ctx context.Context) error {
	if !p.gitHost.Enabled() {
		slog.Warn("Skipping git host because it is not enabled", "githost", p.gitHost.Name())
		return nil
	}

	slog.Info("Starting main workflow...")
	start := time.Now()

	// Get starred repos from the Git provider, we set a limit on concurrent requests so we
	// don't get rate limited by the Git host.
	ghErrgoup, ghCtx := errgroup.WithContext(ctx)
	ghErrgoup.SetLimit(5)
	var starredRepoMap map[string]githost.Repo
	ghErrgoup.Go(func() error {
		repos, err := p.gitHost.GetStarredRepos(ghCtx)
		if err != nil {
			return err
		}
		starredRepoMap = repos

		slog.Info(
			"Queried starred repos in Git host",
			"Git host", p.gitHost.Name(),
			"numberStarredRepos", len(starredRepoMap),
			"duration", time.Since(start),
		)
		return nil
	})

	// Authenticate to FreshRSS
	rssErrgoup, rssCtx := errgroup.WithContext(ctx)
	var filteredRssFeedsMap map[string]rss.RSSFeed
	if p.rssServer.Enabled() {
		if err := p.rssServer.Authenticate(rssCtx); err != nil {
			return err
		}

		// Get existing subscriptions
		rssErrgoup.Go(func() error {
			slog.Info("Querying existing RSS feeds in FreshRSS... ")
			rawRssFeedsMap, err := p.rssServer.GetExistingFeeds(ghCtx)
			if err != nil {
				return err
			}

			// Filter out feeds from the list that are not from this git host
			filteredRssFeedsMap = filterOutNonRepoReleaseFeeds(p.gitHost, rawRssFeedsMap)
			slog.Info(
				"Queried Git project release feeds in FreshRSS",
				"numberFeeds", len(filteredRssFeedsMap),
				"duration", time.Since(start),
			)
			return nil
		})

		// Wait for these two independent operations to finish...
		if err := rssErrgoup.Wait(); err != nil {
			return err
		}
		if err := ghErrgoup.Wait(); err != nil {
			return err
		}

		// We can also overwhelm FreshRSS with this so we will also set a limit
		rssErrgoup, rssCtx = errgroup.WithContext(ctx)
		rssErrgoup.SetLimit(10)
		for _, repo := range starredRepoMap {
			rssErrgoup.Go(func() error {
				return p.publishToFreshRSS(rssCtx, filteredRssFeedsMap, repo)
			})
		}
		for feed := range filteredRssFeedsMap {
			rssErrgoup.Go(func() error {
				return p.removeStaleFeeds(rssCtx, starredRepoMap, feed)
			})
		}

		if err := rssErrgoup.Wait(); err != nil {
			return err
		}

		// Report success
		slog.Info(
			"FreshRSS feeds synced from the Git host successfully",
			"Git host", p.gitHost.Name(),
			"duration", time.Since(start),
		)

	} else {
		slog.Warn("Skipping publishing to rss server because it is not enabled")
	}

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
		slog.Info("Not adding feed as it is already in RSS", "feed", repo.Name())
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
func filterOutNonRepoReleaseFeeds(
	gh githost.GitHost,
	rssFeedMap map[string]rss.RSSFeed,
) map[string]rss.RSSFeed {
	filteredMap := make(map[string]rss.RSSFeed)
	for feedUrl, feedObject := range rssFeedMap {
		if gh.IsReleaseFeed(feedUrl) {
			filteredMap[feedUrl] = feedObject
		} else {
			slog.Debug(
				"Ignoring feeds that are't release feed from a git host so we don't unsubscribe",
				"feed", feedUrl,
			)
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
			"Removing feed from RSS Server as it is no longer starred",
			"feed", rssFeed,
		)
		if err := p.rssServer.RemoveFeed(ctx, rssFeed); err != nil {
			return err
		}
	}
	return nil
}
