package runners

import (
	"context"
	"sync"
)

type MockGitForge struct {
	ExpectedError         error
	ExpectedFeeds         map[string]string
	ExpectedIsReleaseFeed bool
	ExpectedName          string
}

func (m *MockGitForge) LoadFeeds(ctx context.Context) (map[string]string, error) {
	return m.ExpectedFeeds, m.ExpectedError
}

func (m *MockGitForge) Name() string {
	return m.ExpectedName
}

func (m *MockGitForge) IsReleaseFeed(feedURL string) bool {
	return m.ExpectedIsReleaseFeed
}

type MockRssServer struct {
	ExpectedError error
	ExpectedFeeds map[string]struct{}
	ExpectedName  string
	AddedFeeds    []string
	RemovedFeeds  []string
	mu            sync.Mutex
}

func (m *MockRssServer) LoadFeeds(ctx context.Context) (map[string]struct{}, error) {
	return m.ExpectedFeeds, m.ExpectedError
}

func (m *MockRssServer) AddFeed(ctx context.Context, feedURL, name, category string) error {
	if m.ExpectedError == nil {
		m.mu.Lock()
		m.AddedFeeds = append(m.AddedFeeds, feedURL)
		m.mu.Unlock()
	}
	return m.ExpectedError
}

func (m *MockRssServer) RemoveFeed(ctx context.Context, feedURL string) error {
	if m.ExpectedError == nil {
		m.mu.Lock()
		m.RemovedFeeds = append(m.RemovedFeeds, feedURL)
		m.mu.Unlock()
	}
	return m.ExpectedError
}

func (m *MockRssServer) Name() string {
	return m.ExpectedName
}
