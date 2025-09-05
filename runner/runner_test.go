package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/atomicmeganerd/starfeed/github"
	"github.com/atomicmeganerd/starfeed/mocks"
)

const (
	mockGhToken       = "gh_token"
	mockFreshRssUrl   = "http://freshrss.example.com"
	mockFreshRssUser  = "testuser"
	mockFreshRssToken = "freshrss_token"
)

type QueryAndPublishFeedsTestCase struct {
	name      string
	responses []http.Response
	urlRegex  []string
}

func (tc *QueryAndPublishFeedsTestCase) GetTestObject() RepoRSSPublisher {
	mockTransport := mocks.NewMockUrlSelectedRoundTripper(tc.responses, tc.urlRegex)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewRepoRSSPublisher(
		mockGhToken,
		mockFreshRssUrl,
		mockFreshRssUser,
		mockFreshRssToken,
		context.Background(),
		mockClient,
	)
}

func TestQueryAndPublishFeeds(t *testing.T) {
	testCases := []QueryAndPublishFeedsTestCase{
		{
			name: "Successful workflow with no repos",
			responses: []http.Response{
				// FreshRSS auth request
				{
					Body:       io.NopCloser(strings.NewReader(`{"auth_token": "test_token"}`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
				// FreshRSS get existing feeds
				{
					Body:       io.NopCloser(strings.NewReader(`{"feeds": []}`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
				// GitHub get starred repos
				{
					Body:       io.NopCloser(strings.NewReader(`[]`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			urlRegex: []string{
				`.*freshrss.*api.*auth.*`,  // FreshRSS auth
				`.*freshrss.*api.*feeds.*`, // FreshRSS feeds
				`.*api\.github\.com.*`,     // GitHub API
			},
		},
		{
			name: "Authentication failure should exit early",
			responses: []http.Response{
				// FreshRSS auth request fails
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "unauthorized"}`)),
					Status:     "401 Unauthorized",
					StatusCode: http.StatusUnauthorized,
				},
			},
			urlRegex: []string{
				`.*freshrss.*api.*auth.*`, // FreshRSS auth
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisher := tc.GetTestObject()

			// This should not panic or hang
			publisher.QueryAndPublishFeeds()
		})
	}
}

// mock implementations for interface testing
type mockFreshRSSFeedManager struct {
	addFeedCalled    bool
	addFeedError     error
	removeFeedCalled bool
	removeFeedError  error
	addFeedUrl       string
	addFeedName      string
	removeFeedUrl    string
}

func (m *mockFreshRSSFeedManager) Authenticate() error {
	return nil
}

func (m *mockFreshRSSFeedManager) AddFeed(feedUrl, name, category string) error {
	m.addFeedCalled = true
	m.addFeedUrl = feedUrl
	m.addFeedName = name
	return m.addFeedError
}

func (m *mockFreshRSSFeedManager) GetExistingFeeds() (map[string]struct{}, error) {
	return nil, nil
}

func (m *mockFreshRSSFeedManager) RemoveFeed(feedUrl string) error {
	m.removeFeedCalled = true
	m.removeFeedUrl = feedUrl
	return m.removeFeedError
}

type mockAtomFeedChecker struct {
	hasEntries   bool
	checkedFeeds []string
}

func (m *mockAtomFeedChecker) CheckFeedHasEntries(feedUrl string) bool {
	m.checkedFeeds = append(m.checkedFeeds, feedUrl)
	return m.hasEntries
}

func TestPublishToFreshRSS(t *testing.T) {
	testCases := []struct {
		name                string
		existingFeeds       map[string]struct{}
		repo                github.GitHubRepo
		atomHasEntries      bool
		freshRSSAddError    error
		expectedFreshRSSAdd bool
		expectedLogSkip     bool
	}{
		{
			name: "Feed already exists - should skip",
			existingFeeds: map[string]struct{}{
				"https://github.com/user/repo/releases.atom": {},
			},
			repo: github.GitHubRepo{
				Name:    "repo",
				HtmlUrl: "https://github.com/user/repo",
				FeedUrl: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "Feed has no entries - should skip",
			existingFeeds: map[string]struct{}{},
			repo: github.GitHubRepo{
				Name:    "repo",
				HtmlUrl: "https://github.com/user/repo",
				FeedUrl: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      false,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "New feed with entries - should add",
			existingFeeds: map[string]struct{}{},
			repo: github.GitHubRepo{
				Name:    "repo",
				HtmlUrl: "https://github.com/user/repo",
				FeedUrl: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
		{
			name:          "Add feed fails - should handle error gracefully",
			existingFeeds: map[string]struct{}{},
			repo: github.GitHubRepo{
				Name:    "repo",
				HtmlUrl: "https://github.com/user/repo",
				FeedUrl: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			freshRSSAddError:    fmt.Errorf("failed to add feed"),
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mocks
			mockFreshRSS := &mockFreshRSSFeedManager{
				addFeedError: tc.freshRSSAddError,
			}
			mockAtom := &mockAtomFeedChecker{
				hasEntries: tc.atomHasEntries,
			}

			// Create WaitGroup
			var wg sync.WaitGroup
			wg.Add(1)

			// Call function
			go publishToFreshRSS(&wg, mockFreshRSS, mockAtom, tc.existingFeeds, tc.repo)

			// Wait for completion
			wg.Wait()

			// Assert AddFeed was called as expected
			if tc.expectedFreshRSSAdd != mockFreshRSS.addFeedCalled {
				t.Errorf("Expected AddFeed called: %t, got: %t",
					tc.expectedFreshRSSAdd, mockFreshRSS.addFeedCalled)
			}

			// If AddFeed should be called, verify the parameters
			if tc.expectedFreshRSSAdd {
				if mockFreshRSS.addFeedUrl != tc.repo.FeedUrl {
					t.Errorf("Expected AddFeed called with URL %s, got %s",
						tc.repo.FeedUrl, mockFreshRSS.addFeedUrl)
				}
				if mockFreshRSS.addFeedName != tc.repo.Name {
					t.Errorf("Expected AddFeed called with name %s, got %s",
						tc.repo.Name, mockFreshRSS.addFeedName)
				}
			}

			// Check if atom checker was called when expected
			if !tc.expectedLogSkip || tc.atomHasEntries {
				found := false
				for _, checkedFeed := range mockAtom.checkedFeeds {
					if checkedFeed == tc.repo.FeedUrl {
						found = true
						break
					}
				}
				expectedAtomCheck := !tc.expectedLogSkip
				if expectedAtomCheck && !found {
					t.Errorf("Expected AtomFeedChecker to be called for %s", tc.repo.FeedUrl)
				}
			}
		})
	}
}

func TestRemoveStaleFeeds(t *testing.T) {
	testCases := []struct {
		name                   string
		starredRepoMap         map[string]github.GitHubRepo
		rssFeed                string
		freshRSSRemoveError    error
		expectedFreshRSSRemove bool
	}{
		{
			name: "Feed still starred - should not remove",
			starredRepoMap: map[string]github.GitHubRepo{
				"https://github.com/user/repo/releases.atom": {
					Name:    "repo",
					HtmlUrl: "https://github.com/user/repo",
					FeedUrl: "https://github.com/user/repo/releases.atom",
				},
			},
			rssFeed:                "https://github.com/user/repo/releases.atom",
			expectedFreshRSSRemove: false,
		},
		{
			name:                   "Feed no longer starred - should remove",
			starredRepoMap:         map[string]github.GitHubRepo{},
			rssFeed:                "https://github.com/user/old-repo/releases.atom",
			expectedFreshRSSRemove: true,
		},
		{
			name:                   "Remove feed fails - should handle error gracefully",
			starredRepoMap:         map[string]github.GitHubRepo{},
			rssFeed:                "https://github.com/user/old-repo/releases.atom",
			freshRSSRemoveError:    fmt.Errorf("failed to remove feed"),
			expectedFreshRSSRemove: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create mock
			mockFreshRSS := &mockFreshRSSFeedManager{
				removeFeedError: tc.freshRSSRemoveError,
			}

			// Create WaitGroup
			var wg sync.WaitGroup
			wg.Add(1)

			// Call function
			go removeStaleFeeds(&wg, mockFreshRSS, tc.starredRepoMap, tc.rssFeed)

			// Wait for completion
			wg.Wait()

			// Assert RemoveFeed was called as expected
			if tc.expectedFreshRSSRemove != mockFreshRSS.removeFeedCalled {
				t.Errorf("Expected RemoveFeed called: %t, got: %t",
					tc.expectedFreshRSSRemove, mockFreshRSS.removeFeedCalled)
			}

			// If RemoveFeed should be called, verify the parameter
			if tc.expectedFreshRSSRemove {
				if mockFreshRSS.removeFeedUrl != tc.rssFeed {
					t.Errorf("Expected RemoveFeed called with URL %s, got %s",
						tc.rssFeed, mockFreshRSS.removeFeedUrl)
				}
			}
		})
	}
}

func TestNewRepoRSSPublisher(t *testing.T) {
	mockClient := &http.Client{}
	ctx := context.Background()

	publisher := NewRepoRSSPublisher(
		mockGhToken,
		mockFreshRssUrl,
		mockFreshRssUser,
		mockFreshRssToken,
		ctx,
		mockClient,
	)

	if publisher.ghToken != mockGhToken {
		t.Errorf("Expected ghToken %s, got %s", mockGhToken, publisher.ghToken)
	}

	if publisher.freshRssUrl != mockFreshRssUrl {
		t.Errorf("Expected freshRssUrl %s, got %s", mockFreshRssUrl, publisher.freshRssUrl)
	}

	if publisher.freshRssUser != mockFreshRssUser {
		t.Errorf("Expected freshRssUser %s, got %s", mockFreshRssUser, publisher.freshRssUser)
	}

	if publisher.freshRssToken != mockFreshRssToken {
		t.Errorf("Expected freshRssToken %s, got %s", mockFreshRssToken, publisher.freshRssToken)
	}

	if publisher.ctx != ctx {
		t.Error("Expected context to match")
	}

	if publisher.client != mockClient {
		t.Error("Expected client to match")
	}
}

type FilterOutNonGithubFeedsTestCase struct {
	name           string
	inputFeeds     map[string]struct{}
	expectedFeeds  map[string]struct{}
	expectedLength int
}

func TestFilterOutNonGithubFeeds(t *testing.T) {
	testCases := []FilterOutNonGithubFeedsTestCase{
		{
			name: "All feeds are GitHub releases",
			inputFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedLength: 3,
		},
		{
			name: "Mixed GitHub and non-GitHub feeds",
			inputFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://example.com/feed.xml":                {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://blog.example.com/rss":                {},
			},
			expectedFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
			},
			expectedLength: 2,
		},
		{
			name: "No GitHub feeds",
			inputFeeds: map[string]struct{}{
				"https://example.com/feed.xml":  {},
				"https://blog.example.com/rss":  {},
				"https://news.example.com/atom": {},
			},
			expectedFeeds:  map[string]struct{}{},
			expectedLength: 0,
		},
		{
			name:           "Empty input map",
			inputFeeds:     map[string]struct{}{},
			expectedFeeds:  map[string]struct{}{},
			expectedLength: 0,
		},
		{
			name: "GitHub feeds with dots and dashes in names",
			inputFeeds: map[string]struct{}{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
				"https://example.com/feed.xml":                             {},
			},
			expectedFeeds: map[string]struct{}{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
			},
			expectedLength: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create GitHub builder for the test
			mockClient := &http.Client{}
			gh := github.NewGitHubStarredFeedBuilder("token", context.Background(), mockClient)

			// Call the function under test
			result := filterOutNonGithubFeeds(gh, tc.inputFeeds)

			// Check the length
			if len(result) != tc.expectedLength {
				t.Errorf("Expected %d feeds, got %d", tc.expectedLength, len(result))
			}

			// Check each expected feed exists in result
			for expectedFeed := range tc.expectedFeeds {
				if _, exists := result[expectedFeed]; !exists {
					t.Errorf("Expected feed %s to be in result, but it wasn't", expectedFeed)
				}
			}

			// Check no unexpected feeds exist in result
			for resultFeed := range result {
				if _, exists := tc.expectedFeeds[resultFeed]; !exists {
					t.Errorf("Unexpected feed %s found in result", resultFeed)
				}
			}
		})
	}
}
