package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type GitForge interface {
	LoadFeeds(ctx context.Context) (gitforge.StarredRepoMap, error)
}

type RssServer interface {
	LoadFeeds(ctx context.Context, category rss.FeedCategory) (rss.RSSFeedSet, error)
	AddFeed(
		ctx context.Context,
		feedURL common.FeedURL,
		name rss.FeedName,
		category rss.FeedCategory,
	) error
	RemoveFeed(ctx context.Context, feedURL common.FeedURL) error
}
