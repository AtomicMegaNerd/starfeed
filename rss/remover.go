package rss

import (
	"context"
	"log/slog"

	"github.com/atomicmeganerd/starfeed/common"
)

type RSSFeedRemover struct {
	RmChan chan common.FeedURL
	rss    FreshRSS
	logger *slog.Logger
}

func NewRSSFeedRemover(rss FreshRSS, logger *slog.Logger) RSSFeedRemover {
	rmChan := make(chan common.FeedURL)
	return RSSFeedRemover{
		RmChan: rmChan,
		rss:    rss,
		logger: logger,
	}
}

func (r RSSFeedRemover) Init(ctx context.Context) {
	for {
		select {
		case feed := <-r.RmChan:
			if err := r.rss.rmFeed(ctx, feed); err != nil {
				r.logger.Error("Error removing feed from RSS Server", "feed", feed)
			}
		case <-ctx.Done():
			return
		}
	}
}
