package runners

import (
	"context"
	"errors"
	"testing"

	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/testutils"
)

func TestSyncFeeds(t *testing.T) {
	logger := testutils.TestLogger(t)

	testCases := []struct {
		name          string
		gitForge      *MockGitForge
		rssServer     *MockRssServer
		expectAdded   int32
		expectRemoved int32
		expectError   bool
	}{
		{
			name: "success- adds new feeds and removes stale feeds",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{
					"https://github.com/user/new-repo/releases.atom": "new-repo",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: rss.RSSFeedSet{
					"https://github.com/user/old-repo/releases.atom": {},
				},
			},
			expectAdded:   1,
			expectRemoved: 1,
			expectError:   false,
		},
		{
			name: "No feeds to sync",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: rss.RSSFeedSet{},
			},
			expectError: false,
		},
		{
			name: "GitForge LoadFeeds fails",
			gitForge: &MockGitForge{
				ExpectedError: errors.New("failed to load from git forge"),
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: rss.RSSFeedSet{},
			},
			expectError: true,
		},
		{
			name: "RssServer LoadFeeds fails",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedError: errors.New("failed to load from rss server"),
			},
			expectError: true,
		},
		{
			name: "AddFeed fails",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{
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
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: rss.RSSFeedSet{
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
				ExpectedFeeds: gitforge.StarredRepoMap{
					"https://github.com/user/repo1/releases.atom": "repo1",
					"https://github.com/user/repo2/releases.atom": "repo2",
					"https://github.com/user/repo3/releases.atom": "repo3",
					"https://github.com/user/repo4/releases.atom": "repo4",
					"https://github.com/user/repo5/releases.atom": "repo5",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: rss.RSSFeedSet{},
			},
			expectAdded: 5,
			expectError: false,
		},
		{
			name: "Multiple feeds to remove concurrently",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: rss.RSSFeedSet{
					"https://github.com/user/old1/releases.atom": {},
					"https://github.com/user/old2/releases.atom": {},
					"https://github.com/user/old3/releases.atom": {},
					"https://github.com/user/old4/releases.atom": {},
					"https://github.com/user/old5/releases.atom": {},
				},
			},
			expectRemoved: 5,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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

			numAdded := tc.rssServer.NumAdded.Load()
			numRemoved := tc.rssServer.NumRemoved.Load()

			if tc.expectAdded != numAdded {
				t.Fatalf("Expected %d feeds added but added %d", tc.expectAdded, numAdded)
			}

			if tc.expectRemoved != numRemoved {
				t.Fatalf("Expected %d feeds removed but removed %d", tc.expectRemoved, numRemoved)
			}
		})
	}
}
