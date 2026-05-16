package githost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/mocks"
)

const (
	// Repo 1
	repoName1    = "repo1"
	repoHtmlUrl1 = "https://github.com/user/repo1"

	// Repo 2
	repoName2    = "repo2"
	repoHtmlUrl2 = "https://github.com/user/repo2"

	// Repo 3
	repoName3    = "repo3"
	repoHtmlUrl3 = "https://github.com/user/repo3"

	// Repo 4
	repoName4    = "repo4"
	repoHtmlUrl4 = "https://github.com/user/repo4"
)

type GetStarredReposTestCase struct {
	name          string
	responses     []http.Response
	expectedRepos []BaseRepo
	expectError   bool
}

func (tc *GetStarredReposTestCase) GetTestObject() GitHost {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return MockValidGitHub(mockClient)
}

func TestGetStarredRepos(t *testing.T) {
	testCases := []GetStarredReposTestCase{
		{
			name: "Single repo with no pages",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`[
						{
							"name": "` + repoName1 + `",
							"html_url": "` + repoHtmlUrl1 + `"
						}
						]`),
					),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []BaseRepo{
				{
					Kind:     mocks.GitHubType,
					RepoName: repoName1,
					RepoURL:  repoHtmlUrl1,
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
							"html_url": "` + repoHtmlUrl1 + `"
						},
						{
							"name": "` + repoName2 + `",
							"html_url": "` + repoHtmlUrl2 + `"
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
							"html_url": "` + repoHtmlUrl3 + `"
						},
						{
							"name": "` + repoName4 + `",
							"html_url": "` + repoHtmlUrl4 + `"
						}
						]`),
					),
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []BaseRepo{
				{RepoName: repoName1, RepoURL: repoHtmlUrl1, Kind: mocks.GitHubType},
				{RepoName: repoName2, RepoURL: repoHtmlUrl2, Kind: mocks.GitHubType},
				{RepoName: repoName3, RepoURL: repoHtmlUrl3, Kind: mocks.GitHubType},
				{RepoName: repoName4, RepoURL: repoHtmlUrl4, Kind: mocks.GitHubType},
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
			expectedRepos: []BaseRepo{},
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
			expectedRepos: []BaseRepo{},
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
			expectedRepos: []BaseRepo{},
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gh := tc.GetTestObject()
			ctx := context.Background()
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
				repo, ok := repos[expected.FeedURL()]
				if !ok {
					t.Errorf("Expected feed %s not found", expected.FeedURL())
					continue
				}
				if repo.Name() != expected.Name() {
					t.Errorf("Expected name %s, got %s", expected.Name(), repo.Name())
				}
			}
		})
	}
}

type TestIsGitHubRepoTestCase struct {
	name        string
	feedUrl     string
	expectMatch bool
}

func TestIsReleaseFeed(t *testing.T) {
	mockHost := MockValidGitHub(&http.Client{})

	testCases := []TestIsGitHubRepoTestCase{
		{
			name:        "Letters only",
			feedUrl:     "https://github.com/atomicmeganerd/starfeed/releases.atom",
			expectMatch: true,
		},
		{
			name:        "Handle .",
			feedUrl:     "https://github.com/EdenEast/nightfox.nvim/releases.atom",
			expectMatch: true,
		},
		{
			name:        "Handle -",
			feedUrl:     "https://github.com/nix-community/NixOS-WSL/releases.atom",
			expectMatch: true,
		},
		{
			name:        "Handle numbers",
			feedUrl:     "https://github.com/PyO3/pyo3/releases.atom",
			expectMatch: true,
		},
		{
			name:        "Not GitHub",
			feedUrl:     "https://rofl.com/user/repo/releases.atom",
			expectMatch: false,
		},
		{
			name:        "Not release",
			feedUrl:     "https://github.com/atomicmeganerd/starfeed/other.atom",
			expectMatch: false,
		},
	}

	for _, tc := range testCases {
		if tc.expectMatch {
			if !mockHost.IsReleaseFeedForCurrentHost(tc.feedUrl) {
				t.Errorf("Expected feed %s to match but it did not", tc.feedUrl)
			}
		} else {
			if mockHost.IsReleaseFeedForCurrentHost(tc.feedUrl) {
				t.Errorf("Expected feed %s to not match but it did", tc.feedUrl)
			}
		}
	}
}
