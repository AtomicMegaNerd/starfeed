package runners

import (
	"context"

	"github.com/atomicmeganerd/starfeed/mocks"
)

type mockFreshRSS struct {
	addFeedCalled    bool
	addFeedError     error
	removeFeedCalled bool
	removeFeedError  error
	addFeedURL       string
	addFeedName      string
	removeFeedURL    string
}

func (m *mockFreshRSS) Enabled() bool {
	return true
}

func (m *mockFreshRSS) Authenticate(ctx context.Context) error {
	return nil
}

func (m *mockFreshRSS) RSSServerType() string {
	return mocks.FreshRSSType
}

func (m *mockFreshRSS) AddFeed(
	ctx context.Context,
	feedURL, name, category string,
) error {
	m.addFeedCalled = true
	m.addFeedURL = feedURL
	m.addFeedName = name
	return m.addFeedError
}

func (m *mockFreshRSS) GetExistingFeeds(
	ctx context.Context,
) (map[string]struct{}, error) {
	return nil, nil
}

func (m *mockFreshRSS) RemoveFeed(ctx context.Context, feedURL string) error {
	m.removeFeedCalled = true
	m.removeFeedURL = feedURL
	return m.removeFeedError
}
