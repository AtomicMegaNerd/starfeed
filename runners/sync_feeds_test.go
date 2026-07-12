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
	}{}

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
