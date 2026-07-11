package runners

import (
	"context"
	"log/slog"
	"time"

	"github.com/atomicmeganerd/starfeed/common"
	"golang.org/x/sync/errgroup"
)

// RepoRSSPublisher is a struct that manages the main workflow of the application.
type SyncFeedsRunner struct {
	gitForge  GitForge
	rssServer RSSServer
	logger    *slog.Logger
}

func NewSyncFeedsRunner(
	gitForge GitForge,
	rssServer RSSServer,
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

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(10)

	p.addNewReleaseFeeds(ctx, eg)
	existFeeds := p.filterFeedsExistForGitForge(p.rssServer.Feeds())
	p.removeStaleFeeds(ctx, eg, existFeeds)

	if err := eg.Wait(); err != nil {
		return err
	}

	// Report success
	p.logger.Info(
		"RSS feeds synced from the Git forge successfully",
		"duration", time.Since(start),
	)

	return nil
}

func (p SyncFeedsRunner) filterFeedsExistForGitForge(
	feeds common.FeedSet,
) common.FeedSet {
	filteredFeeds := make(common.FeedSet, 0)
	for feed := range feeds {
		if _, exists := p.gitForge.FeedRepoMap()[feed]; exists {
			filteredFeeds[feed] = struct{}{}
		}
	}
	return filteredFeeds
}

func (p SyncFeedsRunner) addNewReleaseFeeds(
	ctx context.Context,
	eg *errgroup.Group,
) {
	for feedURL, repoName := range p.gitForge.FeedRepoMap() {
		eg.Go(func() error {
			return p.rssServer.AddFeed(ctx, feedURL, repoName, p.gitForge.Name())
		})
	}
}

func (p SyncFeedsRunner) removeStaleFeeds(
	ctx context.Context,
	eg *errgroup.Group,
	relFeeds map[string]struct{},
) {
	for feed := range relFeeds {
		eg.Go(func() error {
			p.logger.Info(
				"Removing feed from RSS Server as it is no longer starred",
				"feed",
				feed,
			)
			return p.rssServer.RemoveFeed(ctx, feed)
		})
	}
}
