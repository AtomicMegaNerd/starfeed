package runners

import (
	"context"
	"sync/atomic"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type MockGitForge struct {
	ExpectedLoadError error
	ExpectedFeeds     gitforge.StarredRepoMap
}

func (m *MockGitForge) LoadFeeds(ctx context.Context) (gitforge.StarredRepoMap, error) {
	return m.ExpectedFeeds, m.ExpectedLoadError
}

type MockRssServer struct {
	ExpectedLoadError   error
	ExpectedAddError    error
	ExpectedRemoveError error
	ExpectedFeeds       rss.RSSFeedSet

	// These need to be atomic because we call the real RSS server with multiple goroutines. It
	// has no state to protect but this mock does
	NumAdded   atomic.Int32
	NumRemoved atomic.Int32
}

func (m *MockRssServer) LoadFeeds(
	ctx context.Context, category rss.FeedCategory,
) (rss.RSSFeedSet, error) {
	return m.ExpectedFeeds, m.ExpectedLoadError
}

func (m *MockRssServer) AddFeed(
	ctx context.Context,
	feedURL common.FeedURL,
	name rss.FeedName,
	category rss.FeedCategory,
) error {
	if m.ExpectedAddError == nil {
		m.NumAdded.Add(1)
	}
	return m.ExpectedAddError
}

func (m *MockRssServer) RemoveFeed(ctx context.Context, feedURL common.FeedURL) error {
	if m.ExpectedRemoveError == nil {
		m.NumRemoved.Add(1)
	}
	return m.ExpectedRemoveError
}
