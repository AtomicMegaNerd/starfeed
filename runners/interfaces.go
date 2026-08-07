package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type GitForge interface {
	LoadFeeds(ctx context.Context) (gitforge.FeedResultMap, error)
}

type RssServer interface {
	LoadFeeds(ctx context.Context, category rss.FeedCategory) (*common.Set[common.FeedURL], error)
	AddFeed(
		ctx context.Context,
		feedURL common.FeedURL,
		name rss.FeedName,
		category rss.FeedCategory,
	) error
	RemoveFeed(ctx context.Context, feedURL common.FeedURL) error
}
