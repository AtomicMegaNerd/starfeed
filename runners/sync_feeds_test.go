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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/new-repo/releases.atom": gitforge.GitRepoResult{
						RepoName:          "new-repo",
						RelFeedHasEntries: true,
					},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{},
			},
			rssServer: &MockRssServer{
				ExpectedLoadError: errors.New("failed to load from rss server"),
			},
			expectError: true,
		},
		{
			name: "AddFeed fails but no error",
			gitForge: &MockGitForge{
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/repo/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo",
						RelFeedHasEntries: true,
					},
				},
			},
			rssServer: &MockRssServer{
				ExpectedAddError: errors.New("failed to add feed"),
			},
		},
		{
			name: "RemoveFeed fails",
			gitForge: &MockGitForge{
				ExpectedFeeedResultMap: gitforge.FeedResultMap{},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/old-repo/releases.atom",
				),
				ExpectedRemoveError: errors.New("failed to remove feed"),
			},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/repo1/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo1",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo2/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo2",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo3/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo3",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo4/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo4",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo5/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo5",
						RelFeedHasEntries: true,
					},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/repo1/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo1",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo2/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo2",
						RelFeedHasEntries: true,
					},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/existing/releases.atom": gitforge.GitRepoResult{
						RepoName:          "existing",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/new1/releases.atom": gitforge.GitRepoResult{
						RepoName:          "new1",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/new2/releases.atom": gitforge.GitRepoResult{
						RepoName:          "new2",
						RelFeedHasEntries: true,
					},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/new1/releases.atom": gitforge.GitRepoResult{
						RepoName:          "new1",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/new2/releases.atom": gitforge.GitRepoResult{
						RepoName:          "new2",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/new3/releases.atom": gitforge.GitRepoResult{
						RepoName:          "new3",
						RelFeedHasEntries: true,
					},
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
				ExpectedFeeedResultMap: gitforge.FeedResultMap{},
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
			name: "5xx on release feed must not remove existing feed",
			gitForge: &MockGitForge{
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/repo/releases.atom": gitforge.GitRepoResult{
						RepoName: "repo",
						Err:      common.HTTPError{StatusCode: 500},
					},
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/repo/releases.atom",
				),
			},
			expectAdded:   0,
			expectRemoved: 0,
			expectError:   false,
		},
		{
			name: "Network error on release feed must not remove existing feed",
			gitForge: &MockGitForge{
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/repo/releases.atom": gitforge.GitRepoResult{
						RepoName: "repo",
						Err:      errors.New("connection reset"),
					},
				},
			},
			rssServer: &MockRssServer{
				ExpectedFeeds: common.NewSet[common.FeedURL](
					"https://github.com/user/repo/releases.atom",
				),
			},
			expectAdded:   0,
			expectRemoved: 0,
			expectError:   false,
		},
		{
			name: "Large number of feeds",
			gitForge: &MockGitForge{
				ExpectedFeeedResultMap: gitforge.FeedResultMap{
					"https://github.com/user/repo1/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo1",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo2/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo2",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo3/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo3",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo4/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo4",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo5/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo5",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo6/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo6",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo7/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo7",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo8/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo8",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo9/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo9",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo10/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo10",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo11/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo11",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo12/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo12",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo13/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo13",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo14/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo14",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo15/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo15",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo16/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo16",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo17/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo17",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo18/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo18",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo19/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo19",
						RelFeedHasEntries: true,
					},
					"https://github.com/user/repo20/releases.atom": gitforge.GitRepoResult{
						RepoName:          "repo20",
						RelFeedHasEntries: true,
					},
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
