package github

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

func (tc *GetStarredReposTestCase) GetTestObject() GitHubStarredFeedBuilder {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewGitHubStarredFeedBuilder("mockToken", context.Background(), mockClient)
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
							"html_url": "` + repoHtmlUrl1 + `",
							"releases_url": "` + repoReleasesUrl1 + `"
						}
						]`),
					),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []GitHubRepo{
				{
					Name:    repoName1,
					HtmlUrl: repoHtmlUrl1,
					FeedUrl: repoReleasesUrl1,
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
							"html_url": "` + repoHtmlUrl1 + `",
							"releases_url": "` + repoReleasesUrl1 + `"
						},
						{
							"name": "` + repoName2 + `",
							"html_url": "` + repoHtmlUrl2 + `",
							"releases_url": "` + repoReleasesUrl2 + `"
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
							"html_url": "` + repoHtmlUrl3 + `",
							"releases_url": "` + repoReleasesUrl3 + `"
						},
						{
							"name": "` + repoName4 + `",
							"html_url": "` + repoHtmlUrl4 + `",
							"releases_url": "` + repoReleasesUrl4 + `"
						}
						]`),
					),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			expectedRepos: []GitHubRepo{
				{
					Name:    repoName1,
					HtmlUrl: repoHtmlUrl1,
					FeedUrl: repoReleasesUrl1,
				},
				{
					Name:    repoName2,
					HtmlUrl: repoHtmlUrl2,
					FeedUrl: repoReleasesUrl2,
				},
				{
					Name:    repoName3,
					HtmlUrl: repoHtmlUrl3,

					FeedUrl: repoReleasesUrl3,
				},
				{
					Name:    repoName4,
					HtmlUrl: repoHtmlUrl4,

					FeedUrl: repoReleasesUrl4,
				},
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
		t.Logf("Running test case: %s\n", tc.name)

		gh := tc.GetTestObject()
		repos, err := gh.GetStarredRepos()

		if tc.expectError {
			if err == nil {
				t.Fatalf("Expected an error, got none\n")
			} else {
				continue
			}
		} else {
			if err != nil {
				t.Fatalf("Expected no error, got %v\n", err)
			}
			if len(repos) != len(tc.expectedRepos) {
				t.Fatalf("Expected %d repos, got %d\n", len(tc.expectedRepos), len(repos))
			}
		}
	}
}

type TestIsGithubRepoTestCase struct {
	name        string
	feedUrl     string
	expectMatch bool
}

func TestIsGithubReleaseRepo(t *testing.T) {
	mockClient := http.Client{}
	gh := NewGitHubStarredFeedBuilder("", context.Background(), &mockClient)
	testCases := []TestIsGithubRepoTestCase{
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
			name:        "Not Github",
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
			if !gh.IsGithubReleasesFeed(tc.feedUrl) {
				t.Errorf("Expected feed %s to match but it did not", tc.feedUrl)
			}
		} else {
			if gh.IsGithubReleasesFeed(tc.feedUrl) {
				t.Errorf("Expected feed %s to not match but it did", tc.feedUrl)
			}
		}
	}

}
