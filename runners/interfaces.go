package runners

import (
	"context"
)

type GitForge interface {
	LoadFeeds(ctx context.Context) (map[string]string, error)
	IsReleaseFeed(feedURL string) bool
	Name() string
}

type RssServer interface {
	LoadFeeds(ctx context.Context) (map[string]struct{}, error)
	AddFeed(ctx context.Context, feedURL, name, category string) error
	RemoveFeed(ctx context.Context, feedURL string) error
	Name() string
}
