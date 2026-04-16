package runner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/github"
	"github.com/atomicmeganerd/starfeed/mocks"
	"golang.org/x/sync/errgroup"
)

const (
	mockGhToken       = "gh_token"
	mockFreshRSSURL   = "http://freshrss.example.com"
	mockFreshRSSUser  = "testuser"
	mockFreshRSSToken = "freshrss_token"
)

type QueryAndPublishFeedsTestCase struct {
	name        string
	responses   []http.Response
	urlRegex    []string
	expectError bool
}

func (tc *QueryAndPublishFeedsTestCase) GetTestObject() Runner {
	mockTransport := mocks.NewMockUrlSelectedRoundTripper(tc.responses, tc.urlRegex)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewRepoRSSPublisher(
		&config.Config{
			GitHubToken:   mockGhToken,
			FreshRSSURL:   mockFreshRSSURL,
			FreshRSSUser:  mockFreshRSSUser,
			FreshRSSToken: mockFreshRSSToken,
		},
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
				`.*freshrss.*api.*accounts.*`, // FreshRSS auth
				`.*freshrss.*api.*reader.*`,   // FreshRSS feeds
				`.*api\.github\.com.*`,        // GitHub API
			},
			expectError: false,
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
				`.*freshrss.*api.*accounts.*`, // FreshRSS auth
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisher := tc.GetTestObject()
			ctx := context.Background()
			err := publisher.Run(ctx)

			if tc.expectError == true && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if tc.expectError == false && err != nil {
				t.Fatalf("Unexpected error %q", err)
			}
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
) (map[string]struct{}, error) {
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
				Name:           "repo",
				HTMLURL:        "https://github.com/user/repo",
				ReleaseFeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "Feed has no entries - should skip",
			existingFeeds: map[string]struct{}{},
			repo: github.GitHubRepo{
				Name:           "repo",
				HTMLURL:        "https://github.com/user/repo",
				ReleaseFeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      false,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "New feed with entries - should add",
			existingFeeds: map[string]struct{}{},
			repo: github.GitHubRepo{
				Name:           "repo",
				HTMLURL:        "https://github.com/user/repo",
				ReleaseFeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
		{
			name:          "Add feed fails - should handle error gracefully",
			existingFeeds: map[string]struct{}{},
			repo: github.GitHubRepo{
				Name:           "repo",
				HTMLURL:        "https://github.com/user/repo",
				ReleaseFeedURL: "https://github.com/user/repo/releases.atom",
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

			g := &errgroup.Group{}
			ctx := context.Background()

			// Call function
			g.Go(func() error {
				return publishToFreshRSS(ctx, mockFreshRSS, mockAtom, tc.existingFeeds, tc.repo)
			})

			// Wait for completion and check for errors
			if err := g.Wait(); err != nil {
				if err != tc.freshRSSAddError {
					t.Fatalf("Expected error %v but got %v", tc.freshRSSAddError, err)
				}
			}

			// Assert AddFeed was called as expected
			if tc.expectedFreshRSSAdd != mockFreshRSS.addFeedCalled {
				t.Errorf("Expected AddFeed called: %t, got: %t",
					tc.expectedFreshRSSAdd, mockFreshRSS.addFeedCalled)
			}

			// If AddFeed should be called, verify the parameters
			if tc.expectedFreshRSSAdd {
				if mockFreshRSS.addFeedUrl != tc.repo.ReleaseFeedURL {
					t.Errorf("Expected AddFeed called with URL %s, got %s",
						tc.repo.ReleaseFeedURL, mockFreshRSS.addFeedUrl)
				}
				if mockFreshRSS.addFeedName != tc.repo.Name {
					t.Errorf("Expected AddFeed called with name %s, got %s",
						tc.repo.Name, mockFreshRSS.addFeedName)
				}
			}

			// Check if atom checker was called when expected
			if !tc.expectedLogSkip || tc.atomHasEntries {
				found := slices.Contains(mockAtom.checkedFeeds, tc.repo.ReleaseFeedURL)
				expectedAtomCheck := !tc.expectedLogSkip
				if expectedAtomCheck && !found {
					t.Errorf("Expected AtomFeedChecker to be called for %s", tc.repo.ReleaseFeedURL)
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
					Name:           "repo",
					HTMLURL:        "https://github.com/user/repo",
					ReleaseFeedURL: "https://github.com/user/repo/releases.atom",
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

			g := &errgroup.Group{}
			ctx := context.Background()

			g.Go(func() error {
				return removeStaleFeeds(ctx, mockFreshRSS, tc.starredRepoMap, tc.rssFeed)
			})

			// Wait for completion
			if err := g.Wait(); err != nil {
				if err != tc.freshRSSRemoveError {
					t.Fatalf("Exppected %v but got %v", tc.freshRSSRemoveError, err)
				}
			}

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

	publisher := publishReleasesRunner{
		mockGhToken,
		mockFreshRSSURL,
		mockFreshRSSUser,
		mockFreshRSSToken,
		&config.Config{},
		mockClient,
	}

	if publisher.ghToken != mockGhToken {
		t.Errorf("Expected ghToken %s, got %s", mockGhToken, publisher.ghToken)
	}

	if publisher.freshRSSURL != mockFreshRSSURL {
		t.Errorf("Expected freshRSSUrl %s, got %s", mockFreshRSSURL, publisher.freshRSSURL)
	}

	if publisher.freshRSSUser != mockFreshRSSUser {
		t.Errorf("Expected freshRSSUser %s, got %s", mockFreshRSSUser, publisher.freshRSSUser)
	}

	if publisher.freshRSSToken != mockFreshRSSToken {
		t.Errorf("Expected freshRSSToken %s, got %s", mockFreshRSSToken, publisher.freshRSSToken)
	}

	if publisher.client != mockClient {
		t.Error("Expected client to match")
	}
}

type FilterOutNonGitHubFeedsTestCase struct {
	name           string
	inputFeeds     map[string]struct{}
	expectedFeeds  map[string]struct{}
	expectedLength int
}

func TestFilterOutNonGitHubFeeds(t *testing.T) {
	testCases := []FilterOutNonGitHubFeedsTestCase{
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
			gh := github.NewGitHubStarredFeedBuilder("token", mockClient)

			// Call the function under test
			result := filterOutNonGitHubFeeds(gh, tc.inputFeeds)

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
