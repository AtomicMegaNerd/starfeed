package rss

import (
	"context"
	"log/slog"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
)

type RSSFeedLoader struct {
	LoadChan chan gitforge.GitForgeName
	FeedChan chan common.Set[common.FeedURL]
	rss      FreshRSS
	stop     context.CancelFunc
	logger   *slog.Logger
}

func NewRSSFeedLoader(
	rss FreshRSS,
	stop context.CancelFunc,
	logger *slog.Logger,
) RSSFeedLoader {
	loadChan := make(chan gitforge.GitForgeName, 1)
	feedChan := make(chan common.Set[common.FeedURL], 1)
	return RSSFeedLoader{
		LoadChan: loadChan,
		FeedChan: feedChan,
		rss:      rss,
		stop:     stop,
		logger:   logger,
	}
}

func (l RSSFeedLoader) Init(ctx context.Context) {
	for {
		select {
		case category := <-l.LoadChan:
			if err := l.getFeeds(ctx, category); err != nil {
				l.stop()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// Load all feeds that are under the given category.
func (c RSSFeedLoader) getFeeds(
	ctx context.Context,
	forgeName gitforge.GitForgeName,
) error {
	feedList, err := c.rss.loadFeeds(ctx)
	if err != nil {
		return err
	}

	feedSet := common.NewSet[common.FeedURL]()
	for _, feed := range feedList.Feeds {
		// Only add feeds that are from the category that we care about
		for _, catStruct := range feed.Categories {
			if catStruct.Label == forgeName {
				feedSet.Add(feed.URL)
			}
		}
	}

	c.FeedChan <- *feedSet
	return nil
}
