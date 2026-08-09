package rss

import (
	"context"
	"log/slog"

	"github.com/atomicmeganerd/starfeed/common"
)

type Unsubscriber struct {
	UnsubChan chan common.FeedURL
	rss       RSS
	logger    *slog.Logger
}

func NewUnsubscriber(rss RSS, logger *slog.Logger) Unsubscriber {
	return Unsubscriber{
		UnsubChan: make(chan common.FeedURL),
		rss:       rss,
		logger:    logger,
	}
}

func (r Unsubscriber) Init(ctx context.Context) {
	for {
		select {
		case feed := <-r.UnsubChan:
			if err := r.rss.unsubscribe(ctx, feed); err != nil {
				r.logger.Error("Error removing feed from RSS Server", "feed", feed)
			}
		case <-ctx.Done():
			return
		}
	}
}
