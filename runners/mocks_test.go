package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/common"
)

type MockGitForge struct {
	ExpectedError error
	ExpectedFeeds common.FeedRepoMap
	ExpectedName  string
}

func (m *MockGitForge) FeedRepoMap() common.FeedRepoMap {
	return m.ExpectedFeeds
}

func (m *MockGitForge) Name() string {
	return m.ExpectedName
}

type MockRSSServer struct {
	ExpectedError error
	ExpectedFeeds common.FeedSet
	ExpectedName  string
	AddedFeeds    []string
	RemovedFeeds  []string
}

func (m *MockRSSServer) AddFeed(ctx context.Context, feedURL, name, category string) error {
	m.AddedFeeds = append(m.AddedFeeds, feedURL)
	return m.ExpectedError
}

func (m *MockRSSServer) RemoveFeed(ctx context.Context, feedURL string) error {
	m.RemovedFeeds = append(m.RemovedFeeds, feedURL)
	return m.ExpectedError
}

func (m *MockRSSServer) Feeds() common.FeedSet {
	return m.ExpectedFeeds
}

func (m *MockRSSServer) Name() string {
	return m.ExpectedName
}
