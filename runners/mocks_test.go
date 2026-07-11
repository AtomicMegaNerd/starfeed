package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/gitforge"
)

type MockGitForge struct {
	ExpectedError         error
	ExpectedFeeds         gitforge.FeedRepoMap
	ExpectedIsReleaseFeed bool
	ExpectedEnabled       bool
	ExpectedName          string
}

func (m *MockGitForge) FeedRepoMap() gitforge.FeedRepoMap {
	return m.ExpectedFeeds
}

func (m *MockGitForge) Name() string {
	return m.ExpectedName
}

type MockRSSServer struct {
	ExpectedError   error
	ExpectedFeeds   map[string]struct{}
	ExpectedEnabled bool
	ExpectedName    string
	AddedFeeds      []string
	RemovedFeeds    []string
}

func (m *MockRSSServer) AddFeed(ctx context.Context, feedURL, name, category string) error {
	m.AddedFeeds = append(m.AddedFeeds, feedURL)
	return m.ExpectedError
}

func (m *MockRSSServer) RemoveFeed(ctx context.Context, feedURL string) error {
	m.RemovedFeeds = append(m.RemovedFeeds, feedURL)
	return m.ExpectedError
}

func (m *MockRSSServer) LoadFeeds(ctx context.Context) error {
	return m.ExpectedError
}

func (m *MockRSSServer) Feeds() map[string]struct{} {
	return m.ExpectedFeeds
}

func (m *MockRSSServer) Enabled() bool {
	return m.ExpectedEnabled
}

func (m *MockRSSServer) Name() string {
	return m.ExpectedName
}
