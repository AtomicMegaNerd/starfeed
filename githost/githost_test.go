package githost

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

	"github.com/atomicmeganerd/starfeed/mocks"
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
					Status:     mocks.StatusOKString,
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
				ok := slices.Contains(repos, expected)
				if !ok {
					t.Errorf("Expected repo %s not found", expected.Name)
					continue
				}
			}
		})
	}
}
