package gitforge

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/testutils"
	"github.com/lmittmann/tint"
)

var (
	repo1 = StarredRepo{
		Name:    "repo1",
		RepoURL: "https://github.com/user/repo1",
		FeedURL: "https://github.com/user/repo1/releases.atom",
	}

	repo2 = StarredRepo{
		Name:    "repo2",
		RepoURL: "https://github.com/user/repo2",
		FeedURL: "https://github.com/user/repo2/releases.atom",
	}

	repo3 = StarredRepo{
		Name:    "repo3",
		RepoURL: "https://github.com/user/repo3",
		FeedURL: "https://github.com/user/repo3/releases.atom",
	}

	repo4 = StarredRepo{
		Name:    "repo4",
		RepoURL: "https://github.com/user/repo4",
		FeedURL: "https://github.com/user/repo4/releases.atom",
	}

	MockGitHubConfig = GitForgeConfig{
		Type:  GitHubForgeType,
		Name:  testutils.GitHubName,
		Fqdn:  testutils.GitHubFqdn,
		Token: testutils.GitHubToken,
	}
)

func TestFetchStarredRepos(t *testing.T) {

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
								"name": "` + repo1.Name + `",
								"html_url": "` + repo1.RepoURL + `"
							}
						]`,
						),
					),
					Status:     testutils.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{repo1},
			expectError:   false,
		},
		{
			name: "A few repos over multiple pages",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`[
						{
							"name": "` + repo1.Name + `",
							"html_url": "` + repo1.RepoURL + `"
						},
						{
							"name": "` + repo2.Name + `",
							"html_url": "` + repo2.RepoURL + `"
						}
						]`),
					),
					Status:     testutils.StatusOKString,
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
							"name": "` + repo3.Name + `",
							"html_url": "` + repo3.RepoURL + `"
						},
						{
							"name": "` + repo4.Name + `",
							"html_url": "` + repo4.RepoURL + `"
						}
						]`),
					),
					Status:     testutils.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{repo1, repo2, repo3, repo4},
			expectError:   false,
		},
		{
			name: "404 response should trigger an error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(``)),
					Status:     testutils.StatusNotFoundString,
					StatusCode: http.StatusNotFound,
				},
			},
			expectError: true,
		},
		{
			name: "Reading response body should trigger an error",
			responses: []http.Response{
				{
					Body:       testutils.NewErrorReadCloser(),
					Status:     testutils.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectError: true,
		},
		{
			name: "Invalid json should trigger an error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(testutils.Invalid)),
					Status:     testutils.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}
			gh := NewGitForge(MockGitHubConfig, testutils.TestLogger(), mockClient)

			repos, err := gh.fetchStarredRepos(ctx)

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
				t.Fatalf(
					"Expected %d repos, got %d",
					len(tc.expectedRepos), len(repos),
				)
			}

			for i, repo := range repos {
				expected := tc.expectedRepos[i]
				if repo.Name != expected.Name {
					t.Errorf(
						"Repo %d: expected Name %q, got %q",
						i, expected.Name, repo.Name,
					)
				}
				if repo.RepoURL != expected.RepoURL {
					t.Errorf(
						"Repo %d: expected RepoURL %q, got %q",
						i, expected.RepoURL, repo.RepoURL,
					)
				}
				if repo.FeedURL != expected.FeedURL {
					t.Errorf(
						"Repo %d: expected FeedURL %q, got %q",
						i, expected.FeedURL, repo.FeedURL,
					)
				}
			}
		})
	}
}

func TestCheckReleaseFeedExistsAndHasEntries(t *testing.T) {
	logger := slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
		}),
	)

	testCases := []struct {
		name             string
		repoURL          string
		feedURL          string
		responses        []http.Response
		expectHasEntries bool
		expectError      bool
	}{
		{
			name:    "Feed has entries",
			repoURL: "https://github.com/user/repo1",
			feedURL: "https://github.com/user/repo1/releases.atom",
			responses: []http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`
						<feed xmlns="http://www.w3.org/2005/Atom">
							<entry>
								<title>Entry 1</title>
								<id>1</id>
							</entry>
						</feed>
					`)),
				},
			},
			expectHasEntries: true,
		},
		{
			name:    "Feed has no entries",
			repoURL: "https://github.com/user/repo2",
			feedURL: "https://github.com/user/repo2/releases.atom",
			responses: []http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`
						<feed xmlns="http://www.w3.org/2005/Atom">
						</feed>
					`)),
				},
			},
		},
		{
			name:    "Error making request",
			repoURL: "https://github.com/user/repo3",
			feedURL: "https://github.com/user/repo3/releases.atom",
			responses: []http.Response{
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			},
		},
		{
			name:    "Error reading response",
			repoURL: "https://github.com/user/repo4",
			feedURL: "https://github.com/user/repo4/releases.atom",
			responses: []http.Response{
				{
					Body: testutils.NewErrorReadCloser(),
				},
			},
		},
		{
			name:    "Error parsing XML",
			repoURL: "https://github.com/user/repo5",
			feedURL: "https://github.com/user/repo5/releases.atom",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`
						<feed xmlns="http://www.w3.org/2005/Atom">
							<entry>
								<title>Entry 1</title>
								<id>1</id>
						</feed>
					`)),
				},
			},
		},
		{
			name:    "Not found does not result in error",
			repoURL: "https://github.com/user/repo5",
			feedURL: "https://github.com/user/repo5/releases.atom",
			responses: []http.Response{
				{
					Status:     "Not found",
					StatusCode: http.StatusNotFound,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}
			gh := NewGitForge(MockGitHubConfig, logger, mockClient)

			repo := StarredRepo{RepoURL: tc.repoURL}
			hasEntries := gh.repoHasReleaseFeed(ctx, repo)

			if tc.expectHasEntries != hasEntries {
				t.Fatalf("Expected HasEntries to be %t but got %t", tc.expectHasEntries, hasEntries)
			}
		})
	}
}
