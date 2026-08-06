package runners

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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

	var gitForgeFeeds map[string]string
	var rssServerFeeds map[string]struct{}

	p.logger.Info("Starting publish releases workflow")
	start := time.Now()

	// First load the feeds from each
	eg, loadCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		gitForgeFeeds, err = p.gitForge.LoadFeeds(loadCtx)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func() error {
		var err error
		rssServerFeeds, err = p.rssServer.LoadFeeds(loadCtx)
		if err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error loading the feeds from git or rss: %w", err)
	}

	// Then perform the sync to RSS
	eg, syncCtx := errgroup.WithContext(ctx)
	eg.SetLimit(10)
	p.addNewReleaseFeeds(syncCtx, gitForgeFeeds, eg)
	p.removeStaleFeeds(syncCtx, gitForgeFeeds, rssServerFeeds, eg)
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
	gitForgeFeeds map[string]string,
	eg *errgroup.Group,
) {
	for feedURL, repoName := range gitForgeFeeds {
		eg.Go(func() error {
			return p.rssServer.AddFeed(ctx, feedURL, repoName, p.gitForge.Name())
		})
	}
}

func (p SyncFeedsRunner) removeStaleFeeds(
	ctx context.Context,
	gitForgeFeeds map[string]string,
	rssServerFeeds map[string]struct{},
	eg *errgroup.Group,
) {
	for feed := range rssServerFeeds {
		if p.isRepoFeedStale(gitForgeFeeds, feed) {
			eg.Go(func() error {
				p.logger.Info(
					"Removing feed from RSS Server as it is no longer starred", "feed", feed,
				)
				return p.rssServer.RemoveFeed(ctx, feed)
			})
		}
	}
}

func (p SyncFeedsRunner) isRepoFeedStale(gitForgeFeeds map[string]string, feedUrl string) bool {
	// First of all, if the repo exists it canot be stale
	if _, exists := gitForgeFeeds[feedUrl]; exists {
		return false
	}

	// If the repo does not exist but matches the regex for this gitforge it is stale
	return p.gitForge.IsReleaseFeed(feedUrl)
}
