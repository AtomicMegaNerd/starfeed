package runners

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
	"golang.org/x/sync/errgroup"
)

// RSSServer is an interface that manages the interaction with a FreshRSS instance.
type rssServer interface {
	AddFeed(context.Context, string, string, string) error
	GetExistingFeeds(context.Context) (map[string]struct{}, error)
	RemoveFeed(context.Context, string) error
	Enabled() bool
	RSSServerType() string
}

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type SyncFeedsRunner struct {
	gitHost   githost.GitHost
	rssServer rssServer
	logger    *slog.Logger
}

// NewSyncFeedsRunner creates a new RepoRSSPublisher instance.
// Arguments:
// - gitHost: The git host to query for starred repos.
// - rssServer: The RSS server to publish feeds to.
// - atomFeedChecker: The atom feed checker to verify feed entries.
func NewSyncFeedsRunner(
	gitHost githost.GitHost,
	rssServer rssServer,
	logger *slog.Logger,
) SyncFeedsRunner {
	return SyncFeedsRunner{
		gitHost,
		rssServer,
		logger.With("githost", gitHost.Name, "rsshost", rssServer.RSSServerType()),
	}
}

// This queries release feeds for all starred repos in the specified Git host and publishes them
// to FreshRSS. It also removes any stale release feeds from FreshRSS if they are no longer
// starred.
func (p SyncFeedsRunner) Run(ctx context.Context) error {
	// If this gitHost is not enabled there is nothing to do...
	if !p.gitHost.Enabled {
		p.logger.Warn("Skipping git host because it is not enabled")
		return nil
	}

	p.logger.Info("Starting publish releases workflow")
	start := time.Now()

	// Get starred repos from the Git provider, we set a limit on concurrent requests so we
	// don't get rate limited by the Git host.
	ghErrGroup, ghCtx := errgroup.WithContext(ctx)
	ghErrGroup.SetLimit(5)
	var starredRepoMap map[string]githost.StarredRepo
	ghErrGroup.Go(func() error {
		var err error
		starredRepoMap, err = p.gitHost.GetStarredRepos(ghCtx)
		if err != nil {
			return err
		}

		// Add Release feeds to each repo
		for _, repo := range starredRepoMap {
			if err = p.gitHost.CheckReleaseFeed(ctx, &repo); err != nil {
				return fmt.Errorf(
					"error %w adding release feeds to repo %s from githost %s",
					err, repo.Name, p.gitHost.Name,
				)
			}
		}
		return nil
	})

	// Only publish to RSS if the server is enabled
	if p.rssServer.Enabled() {
		rssErrGroup, rssCtx := errgroup.WithContext(ctx)
		// NOTE: Using map[T]struct{} is idiomatic for creating sets in Go.
		var existingFeeds map[string]struct{}
		var err error
		rssErrGroup.Go(func() error {
			// Get existing subscriptions
			p.logger.Info("Querying existing RSS feeds in FreshRSS... ")
			existingFeeds, err = p.rssServer.GetExistingFeeds(rssCtx)
			if err != nil {
				return err
			}
			p.logger.Info(
				"Queried Git project release feeds in FreshRSS",
				"numberFeeds", len(existingFeeds),
				"duration", time.Since(start),
			)
			return nil
		})

		// Wait for these two independent operations to finish...
		if err := ghErrGroup.Wait(); err != nil {
			return err
		}
		if err := rssErrGroup.Wait(); err != nil {
			return err
		}

		// We can also overwhelm FreshRSS with this so we will also set a limit
		rssErrGroup, rssCtx = errgroup.WithContext(ctx)
		rssErrGroup.SetLimit(10)
		for _, repo := range starredRepoMap {
			rssErrGroup.Go(func() error {
				return p.publishToFreshRSS(rssCtx, existingFeeds, repo)
			})
		}
		for feed := range existingFeeds {
			rssErrGroup.Go(func() error {
				return p.removeStaleFeed(
					rssCtx, starredRepoMap, feed,
				)
			})
		}

		if err := rssErrGroup.Wait(); err != nil {
			return err
		}

		// Report success
		p.logger.Info(
			"FreshRSS feeds synced from the Git host successfully",
			"duration", time.Since(start),
		)

	} else {
		p.logger.Warn("Skipping publishing to rss server because it is not enabled")
		// We also need to wait here for the github queries if the RSS server is disabled
		if err := ghErrGroup.Wait(); err != nil {
			return err
		}
	}

	return nil
}

func (p SyncFeedsRunner) publishToFreshRSS(
	ctx context.Context,
	rssFeedSet map[string]struct{},
	repo githost.StarredRepo,
) error {
	repoFeed := repo.FeedURL
	p.logger.Debug("Checking feed", "repoFeed", repoFeed)

	// If we find that a matching repo in FreshRSS we don't want to add it again...
	if _, exists := rssFeedSet[repoFeed]; exists {
		p.logger.Debug("Not adding feed as it is already in RSS", "feed", repo.Name)
		return nil
	}

	return p.rssServer.AddFeed(ctx, repoFeed, repo.Name, p.gitHost.Name)
}

func (p SyncFeedsRunner) removeStaleFeed(
	ctx context.Context,
	starredRepoMap map[string]githost.StarredRepo, // The key is the release ATOM feed
	rssFeed string,
) error {
	if !p.gitHost.ReleaseFeedPattern.MatchString(rssFeed) {
		p.logger.Debug("Not removing RSS Feed as it is not a release feed for the current githost")
		return nil
	}

	// If a feed does not exist in the Git host, remove it from the RSS server.
	if _, exists := starredRepoMap[rssFeed]; !exists {
		p.logger.Info(
			"Removing feed from RSS Server as it is no longer starred",
			"feed", rssFeed,
		)
		if err := p.rssServer.RemoveFeed(ctx, rssFeed); err != nil {
			return err
		}
	}
	return nil
}
