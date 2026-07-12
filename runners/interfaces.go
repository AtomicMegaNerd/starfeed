package runners

import (
	"context"
)

type GitForge interface {
	LoadFeeds(ctx context.Context) error
	Feeds() map[string]string
	IsRepoFeedStale(feedURL string) bool
	Name() string
}

type RssServer interface {
	LoadFeeds(ctx context.Context) error
	AddFeed(ctx context.Context, feedURL, name, category string) error
	RemoveFeed(ctx context.Context, feedURL string) error
	Feeds() map[string]struct{}
	Name() string
}
