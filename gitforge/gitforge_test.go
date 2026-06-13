package gitforge

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/testutils"
	"github.com/lmittmann/tint"
)

const (
	// Repo 1
	repoName1    = "repo1"
	repoHtmlURL1 = "https://github.com/user/repo1"

	// Repo 2
	repoName2    = "repo2"
	repoHtmlURL2 = "https://github.com/user/repo2"

	// Repo 3
	repoName3    = "repo3"
	repoHtmlURL3 = "https://github.com/user/repo3"

	// Repo 4
	repoName4    = "repo4"
	repoHtmlURL4 = "https://github.com/user/repo4"
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
					Status:     testutils.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{{Name: repoName1, RepoURL: repoHtmlURL1}},
			expectError:   false,
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
							"name": "` + repoName3 + `",
							"html_url": "` + repoHtmlURL3 + `"
						},
						{
							"name": "` + repoName4 + `",
							"html_url": "` + repoHtmlURL4 + `"
						}
						]`),
					),
					Status:     testutils.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []StarredRepo{
				{Name: repoName1, RepoURL: repoHtmlURL1},
				{Name: repoName2, RepoURL: repoHtmlURL2},
				{Name: repoName3, RepoURL: repoHtmlURL3},
				{Name: repoName4, RepoURL: repoHtmlURL4},
			},
			expectError: false,
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
			expectedRepos: []StarredRepo{},
			expectError:   true,
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
			expectedRepos: []StarredRepo{},
			expectError:   true,
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
			expectedRepos: []StarredRepo{},
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockRoundTripper(tc.responses)
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
				ok := slices.Contains(repos, expected)
				if !ok {
					t.Errorf("Expected repo %s not found", expected.Name)
					continue
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
			expectError: false,
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
			expectError: false,
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
			expectError: true,
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
			expectError: true,
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
			expectError: true,
		},
		{
			name:    "Not found does not result in error",
			repoURL: "https://github.com/user/repo5",
			feedURL: "",
			responses: []http.Response{
				{
					Status:     "Not found",
					StatusCode: http.StatusNotFound,
				},
			},
			expectHasEntries: false,
			expectError:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}
			gh := MockValidGitHub(mockClient, logger)

			repo := StarredRepo{RepoURL: tc.repoURL}
			err := gh.CheckReleaseFeedExistsAndHasEntries(ctx, &repo)

			if err != nil && !tc.expectError {
				t.Fatalf("Expected no error, got %v", err)
			}

			if err == nil {
				if tc.expectError {
					t.Fatalf("Expected an error, got none")
				}

				if tc.expectHasEntries {
					if repo.FeedURL != tc.feedURL {
						t.Fatalf("Expected feed %s but got %s", tc.feedURL, repo.FeedURL)
					}
				}
			}
		})
	}
}
