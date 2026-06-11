package runners

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/testutils"
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

func TestSyncFeeds(t *testing.T) {
	logger := testutils.TestLogger()

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
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockURLSelectedRoundTripper(tc.responses, tc.urlRegex)
			mockClient := &http.Client{Transport: &mockTransport}

			// If authentication fails we need to handle that.
			rssServer, err := rss.MockValidRSSEnabledServer(ctx, mockClient, logger)
			// Otherwise test teh rest of the flows
			if err == nil {
				runner := NewSyncFeedsRunner(
					githost.MockValidGitHub(&http.Client{}, logger),
					rssServer,
					logger,
				)

				err = runner.Run(ctx)
			}

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

	exampleRepo := githost.StarredRepo{
		Name:    "repo",
		RepoURL: "https://github.com/user/repo",
		FeedURL: "https://github.com/user/repo/releases.atom",
	}

	exampleRepoJSON, _ := json.Marshal(exampleRepo)
	exampleRepoStr := string(exampleRepoJSON)

	testCases := []struct {
		name             string
		existingFeeds    map[string]struct{}
		repo             githost.StarredRepo
		responses        []http.Response
		atomHasEntries   bool
		expectErr        bool
		expectedAPICalls int
	}{
		{
			name: "Feed already exists - should skip",
			existingFeeds: map[string]struct{}{
				"https://github.com/user/repo/releases.atom": {},
			},
			repo:             exampleRepo,
			atomHasEntries:   true,
			expectedAPICalls: 0,
		},
		{
			name: "Feed has no entries - should skip",
			existingFeeds: map[string]struct{}{
				"https://github.com/user/repo/releases.atom": {},
			},
			repo:             exampleRepo,
			atomHasEntries:   false,
			expectedAPICalls: 0,
		},
		{
			name:          "New feed with entries - should add",
			existingFeeds: map[string]struct{}{},
			repo:          exampleRepo,
			responses: []http.Response{
				{
					// Publish feed
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(exampleRepoStr)),
				},
				{
					// Add to category
					StatusCode: http.StatusOK,
				},
			},
			atomHasEntries:   true,
			expectedAPICalls: 2,
		},
		{
			name:          "Add feed fails - should handle error gracefully",
			existingFeeds: map[string]struct{}{},
			repo: githost.StarredRepo{
				Name:    "repo",
				RepoURL: "https://github.com/user/repo",
				FeedURL: "https://github.com/user/repo/releases.atom",
			},
			responses: []http.Response{
				{
					StatusCode: http.StatusBadRequest,
				},
			},
			atomHasEntries:   true,
			expectErr:        true,
			expectedAPICalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mockTransport := testutils.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}

			mockRunner := &SyncFeedsRunner{
				gitHost:   githost.MockValidGitHub(&http.Client{}, logger),
				rssServer: rss.MockValidRSSServer(ctx, mockClient, logger),
				logger:    logger,
			}

			g := &errgroup.Group{}

			g.Go(func() error {
				return mockRunner.publishToFreshRSS(
					ctx,
					tc.existingFeeds,
					tc.repo,
				)
			})

			err := g.Wait()

			if tc.expectErr {
				if err == nil {
					t.Fatal("Expected error but didn't get one")
				}
			} else {
				if err != nil {
					t.Fatalf("Got an error that we didn't expect %v", err)
				}

				timesCalled := mockTransport.GetNumCalls()
				if timesCalled != tc.expectedAPICalls {
					t.Fatalf(
						"Expected %d API calls but had %d",
						tc.expectedAPICalls,
						timesCalled,
					)
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
		name             string
		starredRepoMap   map[string]githost.StarredRepo
		gitHost          githost.GitHost
		rssFeed          string
		responses        []http.Response
		expectError      bool
		expectedAPICalls int
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
			gitHost: githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed: "https://github.com/user/repo/releases.atom",
		},
		{
			name:           "Github unstarred - should not remove codeberg repo",
			starredRepoMap: map[string]githost.StarredRepo{},
			gitHost:        githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:        "https://codeberg.org/user/repo/releases.atom",
		},
		{
			name:           "Codeberg unstarred - should not remove Github repo",
			starredRepoMap: map[string]githost.StarredRepo{},
			gitHost:        githost.MockValidCodeberg(&http.Client{}, logger),
			rssFeed:        "https://github.com/user/repo/releases.atom",
		},
		{
			name:           "Not a release feed - should not remove",
			starredRepoMap: map[string]githost.StarredRepo{},
			gitHost:        githost.MockValidCodeberg(&http.Client{}, logger),
			rssFeed:        "https://roflstar.com/feed/feed.xml",
		},
		{
			name:           "Feed no longer starred - should remove",
			starredRepoMap: map[string]githost.StarredRepo{},
			gitHost:        githost.MockValidGitHub(&http.Client{}, logger),
			rssFeed:        "https://github.com/user/old-repo/releases.atom",
			responses: []http.Response{
				{
					StatusCode: http.StatusOK,
				},
			},
			expectedAPICalls: 1,
		},
		{
			name:           "Remove feed fails - should handle error gracefully",
			starredRepoMap: map[string]githost.StarredRepo{},
			rssFeed:        "https://github.com/user/old-repo/releases.atom",
			gitHost:        githost.MockValidGitHub(&http.Client{}, logger),
			responses: []http.Response{
				{
					StatusCode: http.StatusBadRequest,
				},
			},
			expectError:      true,
			expectedAPICalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}

			runner := &SyncFeedsRunner{
				rssServer: rss.MockValidRSSServer(ctx, mockClient, logger),
				gitHost:   tc.gitHost,
				logger:    logger,
			}

			g := &errgroup.Group{}
			g.Go(func() error {
				return runner.removeStaleFeed(
					ctx, tc.starredRepoMap, tc.rssFeed,
				)
			})
			err := g.Wait()

			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error but didn't get one")
				}
			} else {
				if err != nil {
					t.Fatalf("Got an error that we didn't expect %v", err)
				}

				timesCalled := mockTransport.GetNumCalls()
				if timesCalled != tc.expectedAPICalls {
					t.Fatalf(
						"Expected %d API calls but had %d",
						tc.expectedAPICalls,
						timesCalled,
					)
				}
			}
		})
	}
}
