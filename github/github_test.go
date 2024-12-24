package github

import (
	"net/http"
	"testing"

	mocks "github.com/atomicmeganerd/starfeed/utils"
)

const (
	fakeToken = "fake-token"
)

type gitHubTestCase struct {
	name          string
	response      string
	expectedRepos map[int64]GitHubRepo
	statusCode    int
	expectError   bool
}

func TestGetStarredRepoTestWorks(t *testing.T) {

	testCases := []gitHubTestCase{
		{
			name: "Successful response should just work",
			response: `
[
  {
    "id": 1296269,
    "name": "Hello-World",
    "full_name": "octocat/Hello-World",
    "owner": {
      "login": "octocat"
    },
    "private": false,
    "html_url": "https://github.com/octocat/Hello-World",
    "description": "This your first repo!",
    "fork": false,
    "url": "https://api.github.com/repos/octocat/Hello-World",
    "releases_url": "https://api.github.com/repos/octocat/Hello-World/releases{/id}"
  }
]
`,
			expectedRepos: map[int64]GitHubRepo{
				1296269: {
					ID:              1296269,
					Name:            "Hello-World",
					HtmlUrl:         "https://github.com/octocat/Hello-World",
					ReleasesFeedUrl: "https://github.com/octocat/Hello-World/releases.atom",
				},
			},
			statusCode:  200,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		mockRoundTripper := mocks.NewMockRoundTripper(tc.response, tc.statusCode, make(http.Header))
		mockClient := &http.Client{
			Transport: &mockRoundTripper,
		}

		gh := NewGitHubStarredFeedBuilder(fakeToken, mockClient)
		repos, err := gh.GetStarredRepos()

		if tc.expectError {
			if err == nil {
				t.Error("expected error but did not get one")
				return
			}
		}

		if err != nil {
			t.Errorf("unexpected error %s", err.Error())
			return
		}

		numRepos := len(repos)
		if numRepos != 1 {
			t.Errorf("expected 1 repo match but got %d", numRepos)
		}

		for _, repo := range repos {
			expectedRepo, ok := tc.expectedRepos[repo.ID]
			if !ok {
				t.Errorf("unexpected repo %d", repo.ID)
				return
			}

			if repo.Name != expectedRepo.Name {
				t.Errorf("expected name %s but got %s", expectedRepo.Name, repo.Name)
			}

			if repo.HtmlUrl != expectedRepo.HtmlUrl {
				t.Errorf("expected releases url %s but got %s", expectedRepo.HtmlUrl, repo.HtmlUrl)
			}

			if repo.ReleasesFeedUrl != expectedRepo.ReleasesFeedUrl {
				t.Errorf("expected releases atom feed %s but got %s", expectedRepo.ReleasesFeedUrl, repo.ReleasesFeedUrl)
			}
		}
	}
}

type NextPageTextCase struct {
	linkHeader   string
	nextPageUrl  string
	responseBody string
}

func TestCheckForNextPage(t *testing.T) {
	testCases := []NextPageTextCase{
		{
			linkHeader: `
<https://api.github.com/repositories/1300192/issues?page=2>; rel="prev", <https://api.github.com/repositories/1300192/issues?page=4>; rel="next", <https://api.github.com/repositories/1300192/issues?page=515>; rel="last", <https://api.github.com/repositories/1300192/issues?page=1>; rel="first"
`,
			nextPageUrl:  "https://api.github.com/repositories/1300192/issues?page=4",
			responseBody: "[]",
		},
	}

	for _, tc := range testCases {
		headers := http.Header{}
		headers.Set("link", tc.linkHeader)

		mockRoundTripper := mocks.NewMockRoundTripper(tc.responseBody, 200, headers)
		mockClient := &http.Client{
			Transport: &mockRoundTripper,
		}

		gh := NewGitHubStarredFeedBuilder(fakeToken, mockClient)
		res, err := gh.client.Get("https://api.github.com/user/starred")
		if err != nil {
			t.Error("unexpected error", err)
		}
		ghResponse, err := gh.processGithubResponse(res)
		if err != nil {
			t.Error("unexpected error", err)
		}
		if ghResponse.nextPage != tc.nextPageUrl {
			t.Errorf("Expected %s, got %s", tc.nextPageUrl, ghResponse.nextPage)
		}
	}
}
