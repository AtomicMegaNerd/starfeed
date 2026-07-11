package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/common"
)

type GitForge interface {
	FeedRepoMap() common.FeedRepoMap
	Name() string
}

type RSSServer interface {
	AddFeed(ctx context.Context, feedURL, name, category string) error
	RemoveFeed(ctx context.Context, feedURL string) error
	Feeds() common.FeedSet
	Name() string
}
