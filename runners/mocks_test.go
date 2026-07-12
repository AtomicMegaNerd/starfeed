package runners

import (
	"context"
)

type MockGitForge struct {
	ExpectedError     error
	ExpectedFeeds     map[string]string
	ExpectedRepoStale bool
	ExpectedName      string
}

func (m *MockGitForge) LoadFeeds(ctx context.Context) error {
	return m.ExpectedError
}

func (m *MockGitForge) Feeds() map[string]string {
	return m.ExpectedFeeds
}

func (m *MockGitForge) Name() string {
	return m.ExpectedName
}

func (m *MockGitForge) IsRepoFeedStale(feedURL string) bool {
	return m.ExpectedRepoStale
}

type MockRssServer struct {
	ExpectedError error
	ExpectedFeeds map[string]struct{}
	ExpectedName  string
	AddedFeeds    []string
	RemovedFeeds  []string
}

func (m *MockRssServer) LoadFeeds(ctx context.Context) error {
	return m.ExpectedError
}

func (m *MockRssServer) AddFeed(ctx context.Context, feedURL, name, category string) error {
	if m.ExpectedError == nil {
		m.AddedFeeds = append(m.AddedFeeds, feedURL)
	}
	return m.ExpectedError
}

func (m *MockRssServer) RemoveFeed(ctx context.Context, feedURL string) error {
	if m.ExpectedError == nil {
		m.RemovedFeeds = append(m.RemovedFeeds, feedURL)
	}
	return m.ExpectedError
}

func (m *MockRssServer) Feeds() map[string]struct{} {
	return m.ExpectedFeeds
}

func (m *MockRssServer) Name() string {
	return m.ExpectedName
}
