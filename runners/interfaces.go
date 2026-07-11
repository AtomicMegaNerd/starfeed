package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/gitforge"
)

type GitForge interface {
	FeedRepoMap() gitforge.FeedRepoMap
	Name() string
}

type RSSServer interface {
	AddFeed(ctx context.Context, feedURL, name, category string) error
	RemoveFeed(ctx context.Context, feedURL string) error
	Feeds() map[string]struct{}
	Name() string
}
