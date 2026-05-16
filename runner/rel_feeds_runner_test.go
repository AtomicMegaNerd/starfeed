package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/mocks"
	"github.com/atomicmeganerd/starfeed/rss"
	"golang.org/x/sync/errgroup"
)

type QueryAndPublishFeedsTestCase struct {
	name        string
	responses   []http.Response
	urlRegex    []string
	expectError bool
}

func (tc *QueryAndPublishFeedsTestCase) GetTestRunner() Runner {
	mockTransport := mocks.NewMockUrlSelectedRoundTripper(tc.responses, tc.urlRegex)
	mockClient := &http.Client{Transport: &mockTransport}
	atomFeedChecker := atom.NewAtomFeedChecker(mockClient)

	return NewPublishReleasesRunner(
		githost.MockValidGitHub(mockClient),
		rss.MockValidRSSServer(mockClient),
		atomFeedChecker,
	)
}

func TestQueryAndPublishFeeds(t *testing.T) {
	testCases := []QueryAndPublishFeedsTestCase{
		{
			name: "Successful workflow with no repos",
			responses: []http.Response{
				// FreshRSS auth request
				{
					Body:       io.NopCloser(strings.NewReader(`Auth=test_token\n`)),
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
				`.*rss.*api.*accounts.*`,
				`.*rss.*api.*reader.*`,
				`.*api\.github\.com.*`,
			},
			expectError: false,
		},
		{
			name: "Authentication failure should exit early",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "unauthorized"}`)),
					Status:     "401 Unauthorized",
					StatusCode: http.StatusUnauthorized,
				},
			},
			urlRegex: []string{
				`.*rss.*api.*accounts.*`,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisher := tc.GetTestRunner()
			ctx := context.Background()
			err := publisher.Run(ctx)

			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error %q", err)
			}
		})
	}
}

type mockFreshRSSFeedManager struct {
	addFeedCalled    bool
	addFeedError     error
	removeFeedCalled bool
	removeFeedError  error
	addFeedUrl       string
	addFeedName      string
	removeFeedUrl    string
}

func (m *mockFreshRSSFeedManager) Authenticate(ctx context.Context) error {
	return nil
}

func (m *mockFreshRSSFeedManager) AddFeed(
	ctx context.Context,
	feedUrl, name, category string,
) error {
	m.addFeedCalled = true
	m.addFeedUrl = feedUrl
	m.addFeedName = name
	return m.addFeedError
}

func (m *mockFreshRSSFeedManager) GetExistingFeeds(
	ctx context.Context,
) (map[string]rss.RSSFeed, error) {
	return nil, nil
}

func (m *mockFreshRSSFeedManager) RemoveFeed(ctx context.Context, feedUrl string) error {
	m.removeFeedCalled = true
	m.removeFeedUrl = feedUrl
	return m.removeFeedError
}

type mockAtomFeedChecker struct {
	hasEntries   bool
	err          error
	checkedFeeds []string
}

func (m *mockAtomFeedChecker) CheckFeedHasEntries(
	ctx context.Context, feedUrl string,
) (bool, error) {
	m.checkedFeeds = append(m.checkedFeeds, feedUrl)
	return m.hasEntries, m.err
}

func TestPublishToFreshRSS(t *testing.T) {
	testCases := []struct {
		name                string
		existingFeeds       map[string]rss.RSSFeed
		repo                githost.Repo
		atomHasEntries      bool
		freshRSSAddError    error
		expectedFreshRSSAdd bool
		expectedLogSkip     bool
	}{
		{
			name: "Feed already exists - should skip",
			existingFeeds: map[string]rss.RSSFeed{
				"https://github.com/user/repo/releases.atom": {},
			},
			repo: &githost.BaseRepo{
				RepoName: "repo",
				RepoURL:  "https://github.com/user/repo",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "Feed has no entries - should skip",
			existingFeeds: map[string]rss.RSSFeed{},
			repo: &githost.BaseRepo{
				RepoName: "repo",
				RepoURL:  "https://github.com/user/repo",
			},
			atomHasEntries:      false,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "New feed with entries - should add",
			existingFeeds: map[string]rss.RSSFeed{},
			repo: &githost.BaseRepo{
				RepoName: "repo",
				RepoURL:  "https://github.com/user/repo",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
		{
			name:          "Add feed fails - should handle error gracefully",
			existingFeeds: map[string]rss.RSSFeed{},
			repo: &githost.BaseRepo{
				RepoName: "repo",
				RepoURL:  "https://github.com/user/repo",
			},
			atomHasEntries:      true,
			freshRSSAddError:    fmt.Errorf("failed to add feed"),
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFreshRSS := &mockFreshRSSFeedManager{
				addFeedError: tc.freshRSSAddError,
			}
			mockAtom := &mockAtomFeedChecker{
				hasEntries: tc.atomHasEntries,
			}

			mockRunner := &publishReleasesRunner{
				rssServer:       mockFreshRSS,
				atomFeedChecker: mockAtom,
			}

			g := &errgroup.Group{}
			ctx := context.Background()

			g.Go(func() error {
				return mockRunner.publishToFreshRSS(
					ctx,
					tc.existingFeeds,
					tc.repo,
				)
			})

			if err := g.Wait(); err != nil {
				if err != tc.freshRSSAddError {
					t.Fatalf("Expected error %v but got %v", tc.freshRSSAddError, err)
				}
			}

			if tc.expectedFreshRSSAdd != mockFreshRSS.addFeedCalled {
				t.Errorf("Expected AddFeed called: %t, got: %t",
					tc.expectedFreshRSSAdd, mockFreshRSS.addFeedCalled)
			}

			if tc.expectedFreshRSSAdd {
				if mockFreshRSS.addFeedUrl != tc.repo.FeedURL() {
					t.Errorf("Expected AddFeed called with URL %s, got %s",
						tc.repo.FeedURL(), mockFreshRSS.addFeedUrl)
				}
				if mockFreshRSS.addFeedName != tc.repo.Name() {
					t.Errorf("Expected AddFeed called with name %s, got %s",
						tc.repo.Name(), mockFreshRSS.addFeedName)
				}
			}

			if !tc.expectedLogSkip || tc.atomHasEntries {
				found := slices.Contains(mockAtom.checkedFeeds, tc.repo.FeedURL())
				expectedAtomCheck := !tc.expectedLogSkip
				if expectedAtomCheck && !found {
					t.Errorf("Expected AtomFeedChecker to be called for %s", tc.repo.FeedURL())
				}
			}
		})
	}
}

func TestRemoveStaleFeeds(t *testing.T) {
	testCases := []struct {
		name                   string
		starredRepoMap         map[string]githost.Repo
		rssFeed                string
		freshRSSRemoveError    error
		expectedFreshRSSRemove bool
	}{
		{
			name: "Feed still starred - should not remove",
			starredRepoMap: map[string]githost.Repo{
				"https://github.com/user/repo/releases.atom": &githost.BaseRepo{
					RepoName: "repo",
					RepoURL:  "https://github.com/user/repo",
				},
			},
			rssFeed:                "https://github.com/user/repo/releases.atom",
			expectedFreshRSSRemove: false,
		},
		{
			name:                   "Feed no longer starred - should remove",
			starredRepoMap:         map[string]githost.Repo{},
			rssFeed:                "https://github.com/user/old-repo/releases.atom",
			expectedFreshRSSRemove: true,
		},
		{
			name:                   "Remove feed fails - should handle error gracefully",
			starredRepoMap:         map[string]githost.Repo{},
			rssFeed:                "https://github.com/user/old-repo/releases.atom",
			freshRSSRemoveError:    fmt.Errorf("failed to remove feed"),
			expectedFreshRSSRemove: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFreshRSS := &mockFreshRSSFeedManager{
				removeFeedError: tc.freshRSSRemoveError,
			}

			mockRunner := &publishReleasesRunner{
				rssServer: mockFreshRSS,
			}

			g := &errgroup.Group{}
			ctx := context.Background()

			g.Go(func() error {
				return mockRunner.removeStaleFeeds(ctx, tc.starredRepoMap, tc.rssFeed)
			})

			if err := g.Wait(); err != nil {
				if err != tc.freshRSSRemoveError {
					t.Fatalf("Expected %v but got %v", tc.freshRSSRemoveError, err)
				}
			}

			if tc.expectedFreshRSSRemove != mockFreshRSS.removeFeedCalled {
				t.Errorf("Expected RemoveFeed called: %t, got: %t",
					tc.expectedFreshRSSRemove, mockFreshRSS.removeFeedCalled)
			}

			if tc.expectedFreshRSSRemove {
				if mockFreshRSS.removeFeedUrl != tc.rssFeed {
					t.Errorf("Expected RemoveFeed called with URL %s, got %s",
						tc.rssFeed, mockFreshRSS.removeFeedUrl)
				}
			}
		})
	}
}

type FilterOutNonGitHubFeedsTestCase struct {
	name           string
	inputFeeds     map[string]rss.RSSFeed
	expectedFeeds  map[string]rss.RSSFeed
	expectedLength int
}

func TestFilterOutNonGitHubFeeds(t *testing.T) {
	testCases := []FilterOutNonGitHubFeedsTestCase{
		{
			name: "All feeds are GitHub releases",
			inputFeeds: map[string]rss.RSSFeed{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedFeeds: map[string]rss.RSSFeed{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedLength: 3,
		},
		{
			name: "Mixed GitHub and non-GitHub feeds",
			inputFeeds: map[string]rss.RSSFeed{
				"https://github.com/user/repo1/releases.atom": {},
				"https://example.com/feed.xml":                {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://blog.example.com/rss":                {},
			},
			expectedFeeds: map[string]rss.RSSFeed{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
			},
			expectedLength: 2,
		},
		{
			name: "No GitHub feeds",
			inputFeeds: map[string]rss.RSSFeed{
				"https://example.com/feed.xml":  {},
				"https://blog.example.com/rss":  {},
				"https://news.example.com/atom": {},
			},
			expectedFeeds:  map[string]rss.RSSFeed{},
			expectedLength: 0,
		},
		{
			name:           "Empty input map",
			inputFeeds:     map[string]rss.RSSFeed{},
			expectedFeeds:  map[string]rss.RSSFeed{},
			expectedLength: 0,
		},
		{
			name: "GitHub feeds with dots and dashes in names",
			inputFeeds: map[string]rss.RSSFeed{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
				"https://example.com/feed.xml":                             {},
			},
			expectedFeeds: map[string]rss.RSSFeed{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
			},
			expectedLength: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gitHost := githost.MockValidGitHub(&http.Client{})
			result := filterOutNonGitHubFeeds(gitHost, tc.inputFeeds)

			if len(result) != tc.expectedLength {
				t.Errorf("Expected %d feeds, got %d", tc.expectedLength, len(result))
			}

			for expectedFeed := range tc.expectedFeeds {
				if _, exists := result[expectedFeed]; !exists {
					t.Errorf("Expected feed %s to be in result, but it wasn't", expectedFeed)
				}
			}

			for resultFeed := range result {
				if _, exists := tc.expectedFeeds[resultFeed]; !exists {
					t.Errorf("Unexpected feed %s found in result", resultFeed)
				}
			}
		})
	}
}
