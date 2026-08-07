package runners

import (
	"context"
	"errors"
	"testing"

	"github.com/atomicmeganerd/starfeed/common"
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
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/old-repo/releases.atom",
				),
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
				ExpectedFeeds: common.NewSet[common.FeedURL](),
			},
			expectError: false,
		},
		{
			name: "GitForge LoadFeeds fails",
			gitForge: &MockGitForge{
				ExpectedLoadError: errors.New("failed to load from git forge"),
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](),
			},
			expectError: true,
		},
		{
			name: "RssServer LoadFeeds fails",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedLoadError: errors.New("failed to load from rss server"),
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
				ExpectedAddError: errors.New("failed to add feed"),
			},
			expectError: true,
		},
		{
			name: "RemoveFeed fails",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/old-repo/releases.atom",
				),
				ExpectedRemoveError: errors.New("failed to remove feed"),
			},
			expectError: true,
		},
		{
			name: "Both LoadFeeds fail simultaneously",
			gitForge: &MockGitForge{
				ExpectedLoadError: errors.New("forge error"),
			},
			rssServer: &MockRssServer{
				ExpectedLoadError: errors.New("rss error"),
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
				ExpectedFeeds: common.NewSet[common.FeedURL](),
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
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/old1/releases.atom",
					"https://github.com/user/old2/releases.atom",
					"https://github.com/user/old3/releases.atom",
					"https://github.com/user/old4/releases.atom",
					"https://github.com/user/old5/releases.atom",
				),
			},
			expectRemoved: 5,
			expectError:   false,
		},
		{
			name: "All feeds already exist - no changes needed",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{
					"https://github.com/user/repo1/releases.atom": "repo1",
					"https://github.com/user/repo2/releases.atom": "repo2",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/repo1/releases.atom",
					"https://github.com/user/repo2/releases.atom",
				),
			},
			expectAdded:   0,
			expectRemoved: 0,
			expectError:   false,
		},
		{
			name: "Mix of adds, removes, and existing feeds",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{
					"https://github.com/user/existing/releases.atom": "existing",
					"https://github.com/user/new1/releases.atom":     "new1",
					"https://github.com/user/new2/releases.atom":     "new2",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/existing/releases.atom",
					"https://github.com/user/stale1/releases.atom",
					"https://github.com/user/stale2/releases.atom",
				),
			},
			expectAdded:   2,
			expectRemoved: 2,
			expectError:   false,
		},
		{
			name: "Only adds - no stale feeds",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{
					"https://github.com/user/new1/releases.atom": "new1",
					"https://github.com/user/new2/releases.atom": "new2",
					"https://github.com/user/new3/releases.atom": "new3",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](),
			},
			expectAdded:   3,
			expectRemoved: 0,
			expectError:   false,
		},
		{
			name: "Only removes - no new feeds",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/stale1/releases.atom",
					"https://github.com/user/stale2/releases.atom",
					"https://github.com/user/stale3/releases.atom",
				),
			},
			expectAdded:   0,
			expectRemoved: 3,
			expectError:   false,
		},
		{
			name: "Large number of feeds",
			gitForge: &MockGitForge{
				ExpectedFeeds: gitforge.StarredRepoMap{
					"https://github.com/user/repo1/releases.atom":  "repo1",
					"https://github.com/user/repo2/releases.atom":  "repo2",
					"https://github.com/user/repo3/releases.atom":  "repo3",
					"https://github.com/user/repo4/releases.atom":  "repo4",
					"https://github.com/user/repo5/releases.atom":  "repo5",
					"https://github.com/user/repo6/releases.atom":  "repo6",
					"https://github.com/user/repo7/releases.atom":  "repo7",
					"https://github.com/user/repo8/releases.atom":  "repo8",
					"https://github.com/user/repo9/releases.atom":  "repo9",
					"https://github.com/user/repo10/releases.atom": "repo10",
					"https://github.com/user/repo11/releases.atom": "repo11",
					"https://github.com/user/repo12/releases.atom": "repo12",
					"https://github.com/user/repo13/releases.atom": "repo13",
					"https://github.com/user/repo14/releases.atom": "repo14",
					"https://github.com/user/repo15/releases.atom": "repo15",
					"https://github.com/user/repo16/releases.atom": "repo16",
					"https://github.com/user/repo17/releases.atom": "repo17",
					"https://github.com/user/repo18/releases.atom": "repo18",
					"https://github.com/user/repo19/releases.atom": "repo19",
					"https://github.com/user/repo20/releases.atom": "repo20",
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](),
			},
			expectAdded:   20,
			expectRemoved: 0,
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			category := rss.FeedCategory(testutils.GitHubName)
			runner := NewSyncFeedsRunner(
				tc.gitForge,
				tc.rssServer,

				category,
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
