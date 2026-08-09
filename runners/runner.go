package runners

import (
	"context"
	"log/slog"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type Runner struct {
	// This channel receives the command to sync
	RunChan chan struct{}

	rssLoader  rss.RSSFeedLoader
	rssAdder   rss.RSSFeedAdder
	rssRemover rss.RSSFeedRemover

	gitLoaders []gitforge.GitForgeLoader
}

func NewRunner(
	ctx context.Context,
	stop context.CancelFunc,
	rssServer rss.FreshRSS,
	gitForges []gitforge.GitForge,
	logger *slog.Logger,
) Runner {
	runChan := make(chan struct{}, 1)

	// Register our RSS components
	rssLoader := rss.NewRSSFeedLoader(rssServer, stop, logger)
	go rssLoader.Init(ctx)
	rssAdder := rss.NewRSSFeedAdder(rssServer, logger)
	go rssAdder.Init(ctx)
	rssRemover := rss.NewRSSFeedRemover(rssServer, logger)
	go rssRemover.Init(ctx)
	gitLoaders := make([]gitforge.GitForgeLoader, len(gitForges))

	// Now register our gitLoader components
	for ix, gitForge := range gitForges {
		gitLoaders[ix] = gitforge.NewGitForgeLoader(gitForge, stop, logger)
		go gitLoaders[ix].Init(ctx)
	}

	return Runner{
		RunChan:    runChan,
		rssLoader:  rssLoader,
		rssAdder:   rssAdder,
		rssRemover: rssRemover,
		gitLoaders: gitLoaders,
	}
}

// This method registers the Runner to listen for messages
func (r Runner) Init(ctx context.Context) {
	for {
		select {
		case <-r.RunChan:
			r.Run()
		case <-ctx.Done():
			return
		}
	}
}

func (r Runner) Run() {
	for _, gitLoader := range r.gitLoaders {
		forgeName := gitLoader.Name
		gitFeeds := common.NewSet[common.FeedURL]()

		// Send the request to get both the rss and git feeds
		r.rssLoader.LoadChan <- forgeName
		gitLoader.LoadChan <- struct{}{}

		// We always get the rssFeeds back as a single slice
		rssFeeds := <-r.rssLoader.FeedChan

		// We get the git feeds one by one
		for gitFeed := range gitLoader.GitFeedsChan {
			if gitFeed.Valid && !rssFeeds.Contains(gitFeed.URL) {
				go r.subscribe(gitFeed.Name, gitFeed.URL, forgeName)
			}

			// Add to our set so we can compare against existing rss feeds
			gitFeeds.Add(gitFeed.URL)
		}

		// Remove any stale feeds that are no longer in FrshRSS
		for rssFeed := range rssFeeds.All() {
			if !gitFeeds.Contains(rssFeed) {
				go r.unsubscribe(rssFeed)
			}
		}
	}
}

func (r Runner) subscribe(
	name gitforge.GitRepoName,
	url common.FeedURL,
	category gitforge.GitForgeName,
) {
	req := rss.AddFeedRequest{
		Name:     name,
		URL:      url,
		Category: category,
	}
	r.rssAdder.AddChan <- req
}

func (r Runner) unsubscribe(rssFeed common.FeedURL) {
	r.rssRemover.RmChan <- rssFeed
}
