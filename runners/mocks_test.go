package runners

import (
	"context"
	"sync/atomic"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
)

type MockGitForge struct {
	ExpectedError error
	ExpectedFeeds gitforge.StarredRepoMap
	ExpectedName  string
}

func (m *MockGitForge) LoadFeeds(ctx context.Context) (gitforge.StarredRepoMap, error) {
	return m.ExpectedFeeds, m.ExpectedError
}

func (m *MockGitForge) Name() string {
	return m.ExpectedName
}

type MockRssServer struct {
	ExpectedError error
	ExpectedFeeds rss.RSSFeedSet
	ExpectedName  string

	// These need to be atomic because we call the real RSS server with multiple goroutines. It
	// has no state to protect but this mock does
	NumAdded   atomic.Int32
	NumRemoved atomic.Int32
}

func (m *MockRssServer) LoadFeeds(
	ctx context.Context, category string,
) (rss.RSSFeedSet, error) {
	return m.ExpectedFeeds, m.ExpectedError
}

func (m *MockRssServer) AddFeed(
	ctx context.Context,
	feedURL common.FeedURL,
	name, category string,
) error {
	if m.ExpectedError == nil {
		m.NumAdded.Add(1)
	}
	return m.ExpectedError
}

func (m *MockRssServer) RemoveFeed(ctx context.Context, feedURL common.FeedURL) error {
	if m.ExpectedError == nil {
		m.NumRemoved.Add(1)
	}
	return m.ExpectedError
}

func (m *MockRssServer) Name() string {
	return m.ExpectedName
}
