package runners

import (
	"context"
	"log/slog"
	"sync"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type Runner struct {
	RunChan    chan struct{}
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
		RunChan:    make(chan struct{}, 1),
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
		relFeeds := common.NewSet[common.FeedURL]()

		// Send the request to get both the rss and git feeds
		wg := sync.WaitGroup{}
		wg.Go(func() { r.rssLoader.LoadChan <- forgeName })
		wg.Go(func() { gitLoader.LoadChan <- struct{}{} })
		wg.Wait()

		// We always get the rssFeeds back as a single slice
		// This should block until we get this back
		rssFeeds := <-r.rssLoader.FeedChan

		// The release feeds from the git forge come in as single messages
		for relFeed := range gitLoader.FeedChan {
			if relFeed.Valid && !rssFeeds.Contains(relFeed.URL) {
				go r.subscribe(relFeed.Name, relFeed.URL, forgeName)
			}
			// Add to our set so we can compare against existing rss feeds
			relFeeds.Add(relFeed.URL)
		}

		// Remove any stale feeds that are no longer in FrshRSS
		for rssFeed := range rssFeeds.All() {
			if !relFeeds.Contains(rssFeed) {
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
