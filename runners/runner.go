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

	rssLoader  rss.Loader
	rssAdder   rss.Subscriber
	rssRemover rss.Unsubscriber

	gitLoaders []gitforge.Loader
}

func NewRunner(
	ctx context.Context,
	stop context.CancelFunc,
	rssServer rss.RSS,
	gitForges []gitforge.GitForge,
	logger *slog.Logger,
) Runner {
	runChan := make(chan struct{}, 1)

	// Register our RSS components
	rssLoader := rss.NewLoader(rssServer, stop, logger)
	go rssLoader.Init(ctx)
	rssAdder := rss.NewSubscriber(rssServer, logger)
	go rssAdder.Init(ctx)
	rssRemover := rss.NewUnsubscriber(rssServer, logger)
	go rssRemover.Init(ctx)
	gitLoaders := make([]gitforge.Loader, len(gitForges))

	// Now register our gitLoader components
	for ix, gitForge := range gitForges {
		gitLoaders[ix] = gitforge.NewLoader(gitForge, stop, logger)
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
		for gitFeed := range gitLoader.FeedChan {
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
	name gitforge.RepoName,
	url common.FeedURL,
	category gitforge.ForgeName,
) {
	req := rss.SubscribeRequest{
		Name:     name,
		URL:      url,
		Category: category,
	}
	r.rssAdder.SubChan <- req
}

func (r Runner) unsubscribe(rssFeed common.FeedURL) {
	r.rssRemover.UnsubChan <- rssFeed
}
