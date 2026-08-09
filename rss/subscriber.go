package rss

import (
	"context"
	"log/slog"
)

type Subscriber struct {
	SubChan chan SubscribeRequest
	rss     RSS
	logger  *slog.Logger
}

func NewSubscriber(
	rss RSS,
	logger *slog.Logger,
) Subscriber {
	return Subscriber{
		SubChan: make(chan SubscribeRequest, 1),
		rss:     rss,
		logger:  logger,
	}
}

func (c Subscriber) Init(ctx context.Context) {
	for {
		select {
		case req := <-c.SubChan:
			if err := c.rss.subscribe(ctx, req); err != nil {
				c.logger.Error("Error subscribing to RSS", "error", err, "feed", req.URL)
			}
		case <-ctx.Done():
			return
		}
	}
}
