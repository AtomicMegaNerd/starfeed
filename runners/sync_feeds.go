package runners

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"golang.org/x/sync/errgroup"
)

// SyncFeedsRunner is a struct that manages the main workflow of the application.
type SyncFeedsRunner struct {
	gitForge  GitForge
	category  rss.FeedCategory
	rssServer RssServer
	logger    *slog.Logger
}

func NewSyncFeedsRunner(
	gitForge GitForge,
	rssServer RssServer,
	category rss.FeedCategory,
	logger *slog.Logger,
) SyncFeedsRunner {
	return SyncFeedsRunner{
		gitForge:  gitForge,
		rssServer: rssServer,
		category:  category,
		logger:    logger,
	}
}

// This queries release feeds for all starred repos in the specified Git host and publishes them
// to FreshRSS. It also removes any stale release feeds from FreshRSS if they are no longer
// starred.
func (r SyncFeedsRunner) Run(ctx context.Context) error {
	var starredFeeds gitforge.StarredRepoMap
	var rssFeeds *common.Set[common.FeedURL]

	r.logger.Info("Starting publish releases workflow")
	start := time.Now()

	// First load the feeds from each source in a separate goroutine for speed
	eg, loadCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		starredFeeds, err = r.gitForge.LoadFeeds(loadCtx)
		if err != nil {
			return fmt.Errorf("error loading feeds from gitforge %s: %w", r.category, err)
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		rssFeeds, err = r.rssServer.LoadFeeds(loadCtx, r.category)
		if err != nil {
			return fmt.Errorf(
				"error loading feeds from rss server from category %s: %w",
				r.category,
				err,
			)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	// Then perform the sync to RSS server adding new release feeds and removing
	// old stale feeds
	eg, syncCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	r.addNewReleaseFeeds(syncCtx, starredFeeds, rssFeeds, eg)
	r.removeStaleFeeds(syncCtx, starredFeeds, rssFeeds, eg)
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error updating feeds: %w", err)
	}

	r.logger.Info(
		"RSS feeds synced from the Git forge successfully",
		"duration", time.Since(start),
	)
	return nil
}

func (r SyncFeedsRunner) addNewReleaseFeeds(
	ctx context.Context,
	starredRepoFeeds gitforge.StarredRepoMap,
	rssServerFeeds *common.Set[common.FeedURL],
	eg *errgroup.Group,
) {
	for feedURL, repoResult := range starredRepoFeeds {
		// Don't add feeds that are already in FreshRSS a second time or do not have entries or
		// querying them failed.
		if rssServerFeeds.Contains(feedURL) || !repoResult.IsOK() {
			continue
		}

		// If the feed is valid spawn a gorountie to subscribe to it
		eg.Go(func() error {
			return r.rssServer.AddFeed(
				ctx,
				feedURL,
				rss.FeedName(repoResult.RepoName.String()),
				r.category,
			)
		})
	}
}

func (r SyncFeedsRunner) removeStaleFeeds(
	ctx context.Context,
	starredRepoFeeds gitforge.StarredRepoMap,
	rssServerFeeds *common.Set[common.FeedURL],
	eg *errgroup.Group,
) {
	// This will only contain the list of feeds that are in the category associated
	// with our GitForge by design. This means we will not delete feeds that have nothing
	// to do with this GitForge.
	for feed := range rssServerFeeds.All() {
		// Get the result for this query if there is one
		repoResult, exists := starredRepoFeeds[feed]

		// If the entry is in the map but we could not query the release feed let us not remove it
		// from FreshRSS. If it is stale we could query the release feed but did not find one.
		// If the result is not Stale it means the feed is still valid or the query failed for some
		// other reason.
		if exists && !repoResult.IsStale() {
			continue
		}

		eg.Go(func() error {
			r.logger.Info(
				"Removing feed from RSS Server as it is no longer starred",
				"feed",
				feed,
			)
			return r.rssServer.RemoveFeed(ctx, feed)
		})
	}
}
