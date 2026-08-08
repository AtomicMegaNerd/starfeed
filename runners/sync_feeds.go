package runners

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"golang.org/x/sync/errgroup"
)

// These interfaces are used to easily mock the concrete objects for GitForge and FreshRSS.
// In the future they may also allow for additional implementations.
type gitForge interface {
	LoadFeeds(ctx context.Context) (gitforge.FeedResultMap, error)
}

type rssServer interface {
	LoadFeeds(ctx context.Context, category rss.FeedCategory) (*common.Set[common.FeedURL], error)
	AddFeed(
		ctx context.Context,
		feedURL common.FeedURL,
		name rss.FeedName,
		category rss.FeedCategory,
	) error
	RemoveFeed(ctx context.Context, feedURL common.FeedURL) error
}

// SyncFeedsRunner is our primary runner orchestration object that does all of the co-ordination
// between the GitForge and the RSS reader to make syncing happen for valid starred repo feeds.
type SyncFeedsRunner struct {
	gitForge  gitForge
	category  rss.FeedCategory
	rssServer rssServer
	logger    *slog.Logger
}

func NewSyncFeedsRunner(
	gitForge gitForge,
	rssServer rssServer,
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
//
// Loading should always work and we will return an error if that fails. Syncing on the other
// hand is best effort. We will log failures but not error out on them. However, it will be
// obvious from the logs if a problem with the server is preventing the adding or removing
// of feeds.
//
// But if loading works in 99% of cases adding/deleting will work as well.
func (r SyncFeedsRunner) Run(ctx context.Context) error {
	var gitForgeFeedResults gitforge.FeedResultMap
	var rssFeeds *common.Set[common.FeedURL]

	start := time.Now()
	r.logger.Info("Starting workflow to sync GiForge release feeds with RSS Server")

	// We fire up separate goroutines in the same errGroup to ensure that all of these loads
	// can happen concurrently
	loadEg, loadCtx := errgroup.WithContext(ctx)
	loadEg.SetLimit(10)
	loadEg.Go(func() error {
		var err error
		gitForgeFeedResults, err = r.gitForge.LoadFeeds(loadCtx)
		if err != nil {
			return fmt.Errorf("error loading feeds from gitforge %s: %w", r.category, err)
		}
		return nil
	})
	loadEg.Go(func() error {
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
	// We block here waiting for all loads to finish
	if err := loadEg.Wait(); err != nil {
		return err
	}

	// Next perform the sync to RSS server adding new release feeds and removing
	// old stale feeds. Here we return the slices of func() error that we can then add to our
	// errgroup.Group
	syncEg := errgroup.Group{}
	syncEg.SetLimit(10)

	// We want to keep track of how many feeds we add and delete
	numAdded := &atomic.Int32{}
	numRemoved := &atomic.Int32{}

	addTasks := r.addNewReleaseFeeds(ctx, gitForgeFeedResults, rssFeeds, numAdded)
	rmTasks := r.removeStaleFeeds(ctx, gitForgeFeedResults, rssFeeds, numRemoved)

	// Fire up our task goroutines
	for _, task := range addTasks {
		syncEg.Go(task)
	}
	for _, task := range rmTasks {
		syncEg.Go(task)
	}

	// We block here waiting for them all to finish
	_ = syncEg.Wait()

	r.logger.Info(
		"Syncing GitForge feeds to RSS completed",
		"duration", time.Since(start),
		"numAdded", int(numAdded.Load()),
		"numRemoved", int(numRemoved.Load()),
	)
	return nil
}

// This method returns a slice of functions that can be ranged over and passed to an
// errgroup.Group for concurrent execution. In this case it will add new feeds when the feed
// does not yet exist in RSS has been validated with IsOK
func (r SyncFeedsRunner) addNewReleaseFeeds(
	ctx context.Context,
	gitForgeFeedResults gitforge.FeedResultMap,
	rssServerFeeds *common.Set[common.FeedURL],
	numAdded *atomic.Int32,
) []func() error {
	tasks := make([]func() error, 0, len(gitForgeFeedResults))
	for feedURL, repoResult := range gitForgeFeedResults {
		// Don't add feeds that are already in FreshRSS a second time or do not have entries or
		// querying them failed.
		if rssServerFeeds.Contains(feedURL) || !repoResult.IsOK() {
			continue
		}
		logger := r.logger.With("feedURL", feedURL)
		// If the feed is valid spawn a task that we can append to the tasks slice
		task := func() error {
			logger.Info("Adding new feed to RSS")
			// Just log on failure for these
			if err := r.rssServer.AddFeed(
				ctx,
				feedURL,
				rss.FeedName(repoResult.RepoName.String()),
				r.category,
			); err != nil {
				logger.Warn("Adding new feed failed", "error", err)
				return nil
			}
			numAdded.Add(1)
			return nil
		}
		tasks = append(tasks, task)
	}
	return tasks
}

// This method returns a slice of functions that can be ranged over and passed to an
// errgroup.Group for concurrent execution. In this case it will remove feeds that are part
// of the GitForge's category but are ether no longer there or are returning a 404.
func (r SyncFeedsRunner) removeStaleFeeds(
	ctx context.Context,
	gitForgeFeedResults gitforge.FeedResultMap,
	rssServerFeeds *common.Set[common.FeedURL],
	numRemoved *atomic.Int32,
) []func() error {
	tasks := make([]func() error, 0, rssServerFeeds.Len())
	// This will only contain the list of feeds that are in the category associated
	// with our GitForge by design. This means we will not delete feeds that have nothing
	// to do with this GitForge.
	for feedURL := range rssServerFeeds.All() {
		// Get the result for this query if there is one
		repoResult, exists := gitForgeFeedResults[feedURL]
		// If the entry is in the map but we could not query the release feed let us not remove it
		// from FreshRSS. If it is stale we could query the release feed but did not find one.
		// If the result is not Stale it means the feed is still valid or the query failed for some
		// other reason.
		if exists && !repoResult.IsStale() {
			continue
		}
		logger := r.logger.With("feedURL", feedURL)
		// If the feed needs to be removed append the task to the tasks slice
		task := func() error {
			logger.Info(
				"Removing feed from RSS Server as it is no longer starred",
			)
			// Just log on failure for these
			if err := r.rssServer.RemoveFeed(ctx, feedURL); err != nil {
				logger.Warn("Removing the stale feed failed", "error", err)
				return nil
			}
			numRemoved.Add(1)
			return nil
		}
		tasks = append(tasks, task)
	}
	return tasks
}
