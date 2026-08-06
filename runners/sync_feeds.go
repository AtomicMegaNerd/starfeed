package runners

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"golang.org/x/sync/errgroup"
)

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type SyncFeedsRunner struct {
	gitForge  GitForge
	rssServer RssServer
	logger    *slog.Logger
}

func NewSyncFeedsRunner(
	gitForge GitForge,
	rssServer RssServer,
	logger *slog.Logger,
) SyncFeedsRunner {
	return SyncFeedsRunner{
		gitForge:  gitForge,
		rssServer: rssServer,
		logger:    logger.With("gitForge", gitForge.Name(), "rsshost", rssServer.Name()),
	}
}

// This queries release feeds for all starred repos in the specified Git host and publishes them
// to FreshRSS. It also removes any stale release feeds from FreshRSS if they are no longer
// starred.
func (p SyncFeedsRunner) Run(ctx context.Context) error {

	var starredFeeds gitforge.StarredRepoMap
	var rssFeeds rss.RSSFeedSet

	p.logger.Info("Starting publish releases workflow")
	start := time.Now()

	// First load the feeds from each source in a separate goroutine for speed
	eg, loadCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		starredFeeds, err = p.gitForge.LoadFeeds(loadCtx)
		if err != nil {
			return fmt.Errorf("error loading feeds from gitforge %s: %w", p.gitForge.Name(), err)
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		rssFeeds, err = p.rssServer.LoadFeeds(loadCtx, p.gitForge.Name())
		if err != nil {
			return fmt.Errorf("error loading feeds from rss server %s: %w", p.rssServer.Name(), err)
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
	p.addNewReleaseFeeds(syncCtx, starredFeeds, rssFeeds, eg)
	p.removeStaleFeeds(syncCtx, starredFeeds, rssFeeds, eg)
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error updating feeds: %w", err)
	}

	p.logger.Info(
		"RSS feeds synced from the Git forge successfully",
		"duration", time.Since(start),
	)
	return nil
}

func (p SyncFeedsRunner) addNewReleaseFeeds(
	ctx context.Context,
	starredRepoFeeds gitforge.StarredRepoMap,
	rssServerFeeds rss.RSSFeedSet,
	eg *errgroup.Group,
) {
	for feedURL, repoName := range starredRepoFeeds {
		// Don't add feeds that are already in FreshRSS a second time
		if _, exists := rssServerFeeds[feedURL]; exists {
			continue
		}
		eg.Go(func() error {
			return p.rssServer.AddFeed(ctx, feedURL, string(repoName), p.gitForge.Name())
		})
	}
}

// Remove feeds that did belong to the current GitForge but are no longer in the list of starred
// release feeds
func (p SyncFeedsRunner) removeStaleFeeds(
	ctx context.Context,
	starredRepoFeeds gitforge.StarredRepoMap,
	rssServerFeeds rss.RSSFeedSet,
	eg *errgroup.Group,
) {
	// This will only contain the list of feeds that are in the category associated
	// with our GitForge by design. This means we will not delete feeds that have nothing
	// to do with this GitForge.
	for feed := range rssServerFeeds {
		// if the feed is still in the gitForge map it is still starred and should not be
		// removed.
		if _, exists := starredRepoFeeds[feed]; exists {
			continue
		}
		eg.Go(func() error {
			p.logger.Info(
				"Removing feed from RSS Server as it is no longer starred",
				"feed",
				feed,
				"gitForge",
				p.gitForge.Name(),
			)
			return p.rssServer.RemoveFeed(ctx, feed)
		})
	}
}
