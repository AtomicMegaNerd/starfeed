package runners

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/testutils"
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

func TestSyncFeeds(t *testing.T) {
	logger := testutils.TestLogger()

	testCases := []struct {
		name        string
		gitForge    GitForge
		rssServer   RssServer
		expectError bool
	}{
		{
			name: "success- adds new feeds and removes stale feeds",
			gitForge: &MockGitForge{
				ExpectedFeeds: map[string]string{
					"https://github.com/user/new-repo/releases.atom": "new-repo",
				},
				ExpectedRepoStale: false,
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: map[string]struct{}{
					"https://github.com/user/old-repo/releases.atom": {},
				},
			},
			expectError: false,
		},
		{
			name: "No feeds to sync",
			gitForge: &MockGitForge{
				ExpectedFeeds: map[string]string{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: map[string]struct{}{},
			},
			expectError: false,
		},
		{
			name: "GitForge LoadFeeds fails",
			gitForge: &MockGitForge{
				ExpectedError: errors.New("failed to load from git forge"),
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: map[string]struct{}{},
			},
			expectError: true,
		},
		{
			name: "RssServer LoadFeeds fails",
			gitForge: &MockGitForge{
				ExpectedFeeds: map[string]string{},
			},
			rssServer: &MockRssServer{
				ExpectedError: errors.New("failed to load from rss server"),
			},
			expectError: true,
		},
		{
			name: "AddFeed fails",
			gitForge: &MockGitForge{
				ExpectedFeeds: map[string]string{
					"https://github.com/user/repo/releases.atom": "repo",
				},
			},
			rssServer: &MockRssServer{
				ExpectedError: errors.New("failed to add feed"),
			},
			expectError: true,
		},
		{
			name: "RemoveFeed fails",
			gitForge: &MockGitForge{
				ExpectedFeeds:     map[string]string{},
				ExpectedRepoStale: true,
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: map[string]struct{}{
					"https://github.com/user/old-repo/releases.atom": {},
				},
				ExpectedError: errors.New("failed to remove feed"),
			},
			expectError: true,
		},
		{
			name: "Both LoadFeeds fail simultaneously",
			gitForge: &MockGitForge{
				ExpectedError: errors.New("forge error"),
			},
			rssServer: &MockRssServer{
				ExpectedError: errors.New("rss error"),
			},
			expectError: true,
		},
		{
			name: "Multiple feeds to add concurrently",
			gitForge: &MockGitForge{
				ExpectedFeeds: map[string]string{
					"https://github.com/user/repo1/releases.atom": "repo1",
					"https://github.com/user/repo2/releases.atom": "repo2",
					"https://github.com/user/repo3/releases.atom": "repo3",
					"https://github.com/user/repo4/releases.atom": "repo4",
					"https://github.com/user/repo5/releases.atom": "repo5",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: map[string]struct{}{},
			},
			expectError: false,
		},
		{
			name: "Multiple feeds to remove concurrently",
			gitForge: &MockGitForge{
				ExpectedFeeds:     map[string]string{},
				ExpectedRepoStale: true,
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: map[string]struct{}{
					"https://github.com/user/old1/releases.atom": {},
					"https://github.com/user/old2/releases.atom": {},
					"https://github.com/user/old3/releases.atom": {},
					"https://github.com/user/old4/releases.atom": {},
					"https://github.com/user/old5/releases.atom": {},
				},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			runner := NewSyncFeedsRunner(
				tc.gitForge,
				tc.rssServer,
				logger,
			)

			err := runner.Run(ctx)

			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error %q", err)
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
		name            string
		forgeType       string
		starredRepoMap  map[string]string
		rssFeed         string
		expectedErr     error
		repoIsStale     bool
		expectedRemoved int
	}{
		{
			name: "Feed still starred - should not remove",
			starredRepoMap: map[string]string{
				"https://github.com/user/repo/releases.atom": "repo",
			},
			rssFeed: "https://github.com/user/repo/releases.atom",
		},
		{
			name:    "Github unstarred - should not remove codeberg repo",
			rssFeed: "https://codeberg.org/user/repo/releases.atom",
		},
		{
			name:    "Codeberg unstarred - should not remove Github repo",
			rssFeed: "https://github.com/user/repo/releases.atom",
		},
		{
			name:    "Not a release feed - should not remove",
			rssFeed: "https://roflstar.com/feed/feed.xml",
		},
		{
			name:            "Feed no longer starred - should remove",
			rssFeed:         "https://github.com/user/old-repo/releases.atom",
			repoIsStale:     true,
			expectedRemoved: 1,
		},
		{
			name:        "Remove feed fails - should handle error gracefully",
			rssFeed:     "https://github.com/user/old-repo/releases.atom",
			repoIsStale: true,
			expectedErr: errors.New("error removing feed"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			rssServer := &MockRssServer{
				ExpectedFeeds: map[string]struct{}{tc.rssFeed: {}},
				ExpectedError: tc.expectedErr,
			}
			gitForge := &MockGitForge{
				ExpectedFeeds:     tc.starredRepoMap,
				ExpectedRepoStale: tc.repoIsStale,
			}

			runner := &SyncFeedsRunner{
				rssServer: rssServer,
				gitForge:  gitForge,
				logger:    logger,
			}

			g := &errgroup.Group{}
			runner.removeStaleFeeds(ctx, g)
			err := g.Wait()

			if tc.expectedErr != nil {
				if err == nil {
					t.Fatal("Expected error but didn't get one")
				}
				return
			}

			if err != nil {
				t.Fatalf("Got an error that we didn't expect %v", err)
				return
			}

			if tc.expectedRemoved != len(rssServer.RemovedFeeds) {
				t.Fatalf(
					"Expected %d feeds to be removed but %d were",
					tc.expectedRemoved,
					len(rssServer.RemovedFeeds),
				)
				return
			}
		})
	}
}
