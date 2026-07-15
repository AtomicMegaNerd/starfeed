package runners

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"
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
		gitForge,
		rssServer,
		logger.With("gitForge", gitForge.Name(), "rsshost", rssServer.Name()),
	}
}

// This queries release feeds for all starred repos in the specified Git host and publishes them
// to FreshRSS. It also removes any stale release feeds from FreshRSS if they are no longer
// starred.
func (p SyncFeedsRunner) Run(ctx context.Context) error {
	p.logger.Info("Starting publish releases workflow")
	start := time.Now()

	// First load the feeds from each
	eg, loadCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return p.gitForge.LoadFeeds(loadCtx)
	})
	eg.Go(func() error {
		return p.rssServer.LoadFeeds(loadCtx)
	})
	if err := eg.Wait(); err != nil {
		return err
	}

	// Then perform the sync to RSS
	eg, syncCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	p.addNewReleaseFeeds(syncCtx, eg)
	p.removeStaleFeeds(syncCtx, eg)
	if err := eg.Wait(); err != nil {
		return err
	}

	p.logger.Info(
		"RSS feeds synced from the Git forge successfully",
		"duration", time.Since(start),
	)
	return nil
}

func (p SyncFeedsRunner) addNewReleaseFeeds(
	ctx context.Context,
	eg *errgroup.Group,
) {
	for feedURL, repoName := range p.gitForge.Feeds() {
		eg.Go(func() error {
			return p.rssServer.AddFeed(ctx, feedURL, repoName, p.gitForge.Name())
		})
	}
}

func (p SyncFeedsRunner) removeStaleFeeds(
	ctx context.Context,
	eg *errgroup.Group,
) {
	for feed := range p.rssServer.Feeds() {
		if p.gitForge.IsRepoFeedStale(feed) {
			eg.Go(func() error {
				p.logger.Info(
					"Removing feed from RSS Server as it is no longer starred", "feed", feed,
				)
				return p.rssServer.RemoveFeed(ctx, feed)
			})
		}
	}
}
