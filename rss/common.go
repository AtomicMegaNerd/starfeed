package rss

import (
	"context"
)

// RSSServer is an interface that manages the interaction with a FreshRSS instance.
type RSSServer interface {
	Authenticate(context.Context) error
	AddFeed(context.Context, string, string, string) error
	GetExistingFeeds(context.Context) (map[string]struct{}, error)
	RemoveFeed(context.Context, string) error
	Enabled() bool
}
