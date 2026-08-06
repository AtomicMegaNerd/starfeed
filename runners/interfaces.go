package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type GitForge interface {
	LoadFeeds(ctx context.Context) (gitforge.StarredRepoMap, error)
	Name() string
}

type RssServer interface {
	LoadFeeds(ctx context.Context, category string) (rss.RSSFeedSet, error)
	AddFeed(ctx context.Context, feedURL common.FeedURL, name, category string) error
	RemoveFeed(ctx context.Context, feedURL common.FeedURL) error
	Name() string
}
