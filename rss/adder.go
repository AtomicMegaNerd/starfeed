package rss

import (
	"context"
	"log/slog"
)

type RSSFeedAdder struct {
	AddChan chan AddFeedRequest
	rss     FreshRSS
	logger  *slog.Logger
}

func NewRSSFeedAdder(
	rss FreshRSS,
	logger *slog.Logger,
) RSSFeedAdder {
	addChan := make(chan AddFeedRequest, 1)
	return RSSFeedAdder{
		AddChan: addChan,
		rss:     rss,
		logger:  logger,
	}
}

func (c RSSFeedAdder) Init(ctx context.Context) {
	for {
		select {
		case req := <-c.AddChan:
			if err := c.rss.addFeed(ctx, req); err != nil {
				c.logger.Error("Error adding feed to RSS", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
