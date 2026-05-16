package github

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/mocks"
)

const (
	// Repo 1
	repoName1        = "repo1"
	repoHtmlUrl1     = "https://github.com/user/repo1"
	repoReleasesUrl1 = "https://github.com/user/repo1/releases.atom"

	// Repo 2
	repoName2        = "repo2"
	repoHtmlUrl2     = "https://github.com/user/repo2"
	repoReleasesUrl2 = "https://github.com/user/repo2/releases.atom"

	// Repo 3
	repoName3        = "repo3"
	repoHtmlUrl3     = "https://github.com/user/repo3"
	repoReleasesUrl3 = "https://github.com/user/repo3/releases.atom"

	// Repo 4
	repoName4        = "repo4"
	repoHtmlUrl4     = "https://github.com/user/repo4"
	repoReleasesUrl4 = "https://github.com/user/repo4/releases.atom"
)

type GetStarredReposTestCase struct {
	name          string
	responses     []http.Response
	expectedRepos []GitHubRepo
	expectError   bool
}

func (tc *GetStarredReposTestCase) GetTestObject() githost.GitHost {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	gitHost := githost.GitHostConfig{
		Type:    githost.GitHub,
		BaseURL: "https://github.com",
		Token:   "test_token",
	}
	return NewGitHubStarredFeedBuilder(gitHost, mockClient)
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
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []GitHubRepo{
				{
					RepoName: repoName1,
					HTMLURL:  repoHtmlUrl1,
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
					Status:     "200 OK",
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
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []GitHubRepo{
				{RepoName: repoName1, HTMLURL: repoHtmlUrl1},
				{RepoName: repoName2, HTMLURL: repoHtmlUrl2},
				{RepoName: repoName3, HTMLURL: repoHtmlUrl3},
				{RepoName: repoName4, HTMLURL: repoHtmlUrl4},
			},
			expectError: false,
		},
		{
			name: "404 response should trigger an error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(``)),
					Status:     "404 Not Found",
					StatusCode: http.StatusNotFound,
				},
			},
			expectedRepos: []GitHubRepo{},
			expectError:   true,
		},
		{
			name: "Reading response body should trigger an error",
			responses: []http.Response{
				{
					Body:       mocks.NewErrorReadCloser(),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []GitHubRepo{},
			expectError:   true,
		},
		{
			name: "Invalid json should trigger an error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`The higher, the fewer`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []GitHubRepo{},
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
	mockClient := http.Client{}
	gitHost := githost.GitHostConfig{
		Type:    githost.GitHub,
		BaseURL: "https://github.com",
		Token:   "test_token",
	}
	gh := NewGitHubStarredFeedBuilder(gitHost, &mockClient)
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
			if !gh.IsReleaseFeed(tc.feedUrl) {
				t.Errorf("Expected feed %s to match but it did not", tc.feedUrl)
			}
		} else {
			if gh.IsReleaseFeed(tc.feedUrl) {
				t.Errorf("Expected feed %s to not match but it did", tc.feedUrl)
			}
		}
	}
}
