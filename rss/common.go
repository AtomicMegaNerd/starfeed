package rss

import (
	"context"
)

// RSSServer is an interface that manages the interaction with a FreshRSS instance.
type RSSServer interface {
	Authenticate(ctx context.Context) error
	AddFeed(ctx context.Context, feedUrl, name, category string) error
	GetExistingFeeds(ctx context.Context) (map[string]RSSFeed, error)
	RemoveFeed(ctx context.Context, feedUrl string) error
	Enabled() bool
}
