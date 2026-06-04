package githost

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/mocks"
	"github.com/lmittmann/tint"
)

const (
	// Repo 1
	repoName1    = "repo1"
	repoHtmlURL1 = "https://github.com/user/repo1"
	repoFeedURL1 = "https://github.com/user/repo1/releases.atom"

	// Repo 2
	repoName2    = "repo2"
	repoHtmlURL2 = "https://github.com/user/repo2"
	repoFeedURL2 = "https://github.com/user/repo2/releases.atom"

	// Repo 3
	repoName3    = "repo3"
	repoHtmlURL3 = "https://github.com/user/repo3"
	repoFeedURL3 = "https://github.com/user/repo3/releases.atom"

	// Repo 4
	repoName4    = "repo4"
	repoHtmlURL4 = "https://github.com/user/repo4"
	repoFeedURL4 = "https://github.com/user/repo4/releases.atom"
)

func TestGetStarredRepos(t *testing.T) {
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	testCases := []struct {
		name          string
		responses     []http.Response
		expectedRepos []StarredRepo
		expectError   bool
	}{
		{
			name: "Single repo with no pages",
			responses: []http.Response{
				{
					Body: io.NopCloser(
						strings.NewReader(`[
							{
								"name": "` + repoName1 + `",
								"html_url": "` + repoHtmlURL1 + `"
							}
						]`,
						),
					),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{
				{
					Name:    repoName1,
					RepoURL: repoHtmlURL1,
					FeedURL: repoFeedURL1,
				},
			},
			expectError: false,
		},
		{
			name: "A few repos over multiple pages",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`[
						{
							"name": "` + repoName1 + `",
							"html_url": "` + repoHtmlURL1 + `"
						},
						{
							"name": "` + repoName2 + `",
							"html_url": "` + repoHtmlURL2 + `"
						}
						]`),
					),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Link": []string{
							`<https://api.github.com/user/starred?per_page=2&page=2>; rel="next"`,
						},
					},
				},
				{
					Body: io.NopCloser(strings.NewReader(`[
						{
							"name": "` + repoName3 + `",
							"html_url": "` + repoHtmlURL3 + `"
						},
						{
							"name": "` + repoName4 + `",
							"html_url": "` + repoHtmlURL4 + `"
						}
						]`),
					),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{
				{Name: repoName1, RepoURL: repoHtmlURL1, FeedURL: repoFeedURL1},
				{Name: repoName2, RepoURL: repoHtmlURL2, FeedURL: repoFeedURL2},
				{Name: repoName3, RepoURL: repoHtmlURL3, FeedURL: repoFeedURL3},
				{Name: repoName4, RepoURL: repoHtmlURL4, FeedURL: repoFeedURL4},
			},
			expectError: false,
		},
		{
			name: "404 response should trigger an error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(``)),
					Status:     mocks.StatusNotFoundString,
					StatusCode: http.StatusNotFound,
				},
			},
			expectedRepos: []StarredRepo{},
			expectError:   true,
		},
		{
			name: "Reading response body should trigger an error",
			responses: []http.Response{
				{
					Body:       mocks.NewErrorReadCloser(),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{},
			expectError:   true,
		},
		{
			name: "Invalid json should trigger an error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(mocks.Invalid)),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{},
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockTransport := mocks.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}
			gh := MockValidGitHub(mockClient, logger)

			repos, err := gh.GetStarredRepos(ctx)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected an error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(repos) != len(tc.expectedRepos) {
				t.Fatalf("Expected %d repos, got %d", len(tc.expectedRepos), len(repos))
			}

			for _, expected := range tc.expectedRepos {
				repo, ok := repos[expected.FeedURL]
				if !ok {
					t.Errorf("Expected feed %s not found", expected.FeedURL)
					continue
				}
				if repo.Name != expected.Name {
					t.Errorf("Expected name %s, got %s", expected.Name, repo.Name)
				}
			}
		})
	}
}

func TestFilterOutNonGitHubFeeds(t *testing.T) {
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	testCases := []struct {
		name           string
		inputFeeds     map[string]StarredRepo
		expectedFeeds  map[string]StarredRepo
		expectedLength int
	}{
		{
			name: "All feeds are GitHub releases",
			inputFeeds: map[string]StarredRepo{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedFeeds: map[string]StarredRepo{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedLength: 3,
		},
		{
			name: "Mixed GitHub and non-GitHub feeds",
			inputFeeds: map[string]StarredRepo{
				"https://github.com/user/repo1/releases.atom": {},
				"https://example.com/feed.xml":                {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://blog.example.com/rss":                {},
			},
			expectedFeeds: map[string]StarredRepo{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
			},
			expectedLength: 2,
		},
		{
			name: "No GitHub feeds",
			inputFeeds: map[string]StarredRepo{
				"https://example.com/feed.xml":  {},
				"https://blog.example.com/rss":  {},
				"https://news.example.com/atom": {},
			},
			expectedFeeds:  map[string]StarredRepo{},
			expectedLength: 0,
		},
		{
			name:           "Empty input map",
			inputFeeds:     map[string]StarredRepo{},
			expectedFeeds:  map[string]StarredRepo{},
			expectedLength: 0,
		},
		{
			name: "GitHub feeds with dots and dashes in names",
			inputFeeds: map[string]StarredRepo{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
				"https://example.com/feed.xml":                             {},
			},
			expectedFeeds: map[string]StarredRepo{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
			},
			expectedLength: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gh := MockValidGitHub(&http.Client{}, logger)
			result := gh.filterOutNonRepoReleaseFeeds(tc.inputFeeds)

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
