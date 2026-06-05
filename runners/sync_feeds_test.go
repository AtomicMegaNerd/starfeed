package runners

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/mocks"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

func TestSyncFeeds(t *testing.T) {
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	testCases := []struct {
		name        string
		responses   []http.Response
		urlRegex    []string
		expectError bool
	}{
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
				`.*api\.[a-z0-9]*\.com.*`,
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
			ctx := context.Background()
			mockTransport := mocks.NewMockURLSelectedRoundTripper(tc.responses, tc.urlRegex)
			mockClient := &http.Client{Transport: &mockTransport}

			publisher := NewSyncFeedsRunner(
				githost.MockValidGitHub(mockClient, logger),
				rss.MockValidRSSServer(mockClient, logger),
				mocks.TestLogger(),
			)
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

func TestPublishToFreshRSS(t *testing.T) {
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	testCases := []struct {
		name                string
		existingFeeds       map[string]struct{}
		repo                githost.StarredRepo
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
			repo: githost.StarredRepo{
				Name:    "repo",
				RepoURL: "https://github.com/user/repo",
				FeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name: "Feed has no entries - should skip",
			existingFeeds: map[string]struct{}{
				"https://github.com/user/repo/releases.atom": {},
			},
			repo: githost.StarredRepo{
				Name:    "repo",
				RepoURL: "https://github.com/user/repo",
				FeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      false,
			expectedFreshRSSAdd: false,
			expectedLogSkip:     true,
		},
		{
			name:          "New feed with entries - should add",
			existingFeeds: map[string]struct{}{},
			repo: githost.StarredRepo{
				Name:    "repo",
				RepoURL: "https://github.com/user/repo",
				FeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
		{
			name:          "Add feed fails - should handle error gracefully",
			existingFeeds: map[string]struct{}{},
			repo: githost.StarredRepo{
				Name:    "repo",
				RepoURL: "https://github.com/user/repo",
				FeedURL: "https://github.com/user/repo/releases.atom",
			},
			atomHasEntries:      true,
			freshRSSAddError:    fmt.Errorf("failed to add feed"),
			expectedFreshRSSAdd: true,
			expectedLogSkip:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFreshRSS := &mockFreshRSS{
				addFeedError: tc.freshRSSAddError,
			}

			mockRunner := &SyncFeedsRunner{
				gitHost:   githost.MockValidGitHub(&http.Client{}, logger),
				rssServer: mockFreshRSS,
				logger:    logger,
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
				t.Errorf("Expected call to AddFeed: %t, got: %t",
					tc.expectedFreshRSSAdd, mockFreshRSS.addFeedCalled)
			}

			if tc.expectedFreshRSSAdd {
				if mockFreshRSS.addFeedURL != tc.repo.FeedURL {
					t.Errorf("Expected call toAddFeed with URL %s, got %s",
						tc.repo.FeedURL, mockFreshRSS.addFeedURL)
				}
				if mockFreshRSS.addFeedName != tc.repo.Name {
					t.Errorf("Expected call to AddFeed with name %s, got %s",
						tc.repo.Name, mockFreshRSS.addFeedName)
				}
			}
		})
	}
}

func TestRemoveStaleFeed(t *testing.T) {
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	testCases := []struct {
		name                   string
		starredRepoMap         map[string]githost.StarredRepo
		mockGitHost            githost.GitHost
		rssFeed                string
		freshRSSRemoveError    error
		expectedFreshRSSRemove bool
	}{
		{
			name: "Feed still starred - should not remove",
			starredRepoMap: map[string]githost.StarredRepo{
				"https://github.com/user/repo/releases.atom": {
					Name:    "repo",
					RepoURL: "https://github.com/user/repo",
					FeedURL: "https://github.com/user/repo/releases.atom",
				},
			},
			mockGitHost:            githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:                "https://github.com/user/repo/releases.atom",
			expectedFreshRSSRemove: false,
		},
		{
			name:                   "Github unstarred - should not remove codeberg repo",
			starredRepoMap:         map[string]githost.StarredRepo{},
			mockGitHost:            githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:                "https://codeberg.org/user/repo/releases.atom",
			expectedFreshRSSRemove: false,
		},
		{
			name:                   "Codeberg unstarred - should not remove Github repo",
			starredRepoMap:         map[string]githost.StarredRepo{},
			mockGitHost:            githost.MockValidCodeberg(&http.Client{}, logger),
			rssFeed:                "https://github.com/user/repo/releases.atom",
			expectedFreshRSSRemove: false,
		},
		{
			name:                   "Not a release feed - should no remove",
			starredRepoMap:         map[string]githost.StarredRepo{},
			mockGitHost:            githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:                "https://roflstar.com/feed/feed.xml",
			expectedFreshRSSRemove: false,
		},
		{
			name:                   "Feed no longer starred - should remove",
			starredRepoMap:         map[string]githost.StarredRepo{},
			mockGitHost:            githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:                "https://github.com/user/old-repo/releases.atom",
			expectedFreshRSSRemove: true,
		},
		{
			name:                   "Remove feed fails - should handle error gracefully",
			starredRepoMap:         map[string]githost.StarredRepo{},
			mockGitHost:            githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:                "https://github.com/user/old-repo/releases.atom",
			freshRSSRemoveError:    fmt.Errorf("failed to remove feed"),
			expectedFreshRSSRemove: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFreshRSS := &mockFreshRSS{
				removeFeedError: tc.freshRSSRemoveError,
			}

			mockRunner := &SyncFeedsRunner{
				gitHost:   tc.mockGitHost,
				rssServer: mockFreshRSS,
				logger:    mocks.TestLogger(),
			}

			g := &errgroup.Group{}
			ctx := context.Background()

			g.Go(func() error {
				return mockRunner.removeStaleFeed(
					ctx, tc.starredRepoMap, tc.rssFeed,
				)
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
				if mockFreshRSS.removeFeedURL != tc.rssFeed {
					t.Errorf("Expected RemoveFeed called with URL %s, got %s",
						tc.rssFeed, mockFreshRSS.removeFeedURL)
				}
			}
		})
	}
}
