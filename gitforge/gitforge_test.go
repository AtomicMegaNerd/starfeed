package gitforge

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/testutils"
)

var (
	repo1 = GitRepo{
		Name:    "repo1",
		RepoURL: "https://github.com/user/repo1",
		FeedURL: "https://github.com/user/repo1/releases.atom",
	}

	repo2 = GitRepo{
		Name:    "repo2",
		RepoURL: "https://github.com/user/repo2",
		FeedURL: "https://github.com/user/repo2/releases.atom",
	}
)

func TestLoadFeeds(t *testing.T) {
	testCases := []struct {
		name          string
		mocks         []testutils.MockRoutedResponse
		expectedFeeds FeedResultMap
	}{
		{
			name: "Single repo with valid feed",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred`,
					Response: http.Response{
						Body: io.NopCloser(
							strings.NewReader(`[
								{
									"name": "` + repo1.Name.String() + `",
									"html_url": "` + repo1.RepoURL.String() + `"
								}
							]`,
							),
						),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
				{
					UrlPattern: repo1.Name.String() + `/releases\.atom`,
					Response: http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`
							<feed xmlns="http://www.w3.org/2005/Atom">
								<entry>
									<title>Release 1</title>
									<id>1</id>
								</entry>
							</feed>
						`)),
					},
				},
			},
			expectedFeeds: FeedResultMap{
				repo1.FeedURL: GitRepoResult{
					RepoName:          repo1.Name,
					RelFeedHasEntries: true,
				},
			},
		},
		{
			name: "Multiple repos with mixed feed states",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred`,
					Response: http.Response{
						Body: io.NopCloser(strings.NewReader(`[
							{
								"name": "` + repo1.Name.String() + `",
								"html_url": "` + repo1.RepoURL.String() + `"
							},
							{
								"name": "` + repo2.Name.String() + `",
								"html_url": "` + repo2.RepoURL.String() + `"
							}
							]`),
						),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
				{
					UrlPattern: repo1.Name.String() + `/releases\.atom`,
					Response: http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`
							<feed xmlns="http://www.w3.org/2005/Atom">
								<entry>
									<title>Release 1</title>
									<id>1</id>
								</entry>
							</feed>
						`)),
					},
				},
				{
					UrlPattern: repo2.Name.String() + `/releases\.atom`,
					Response: http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`
							<feed xmlns="http://www.w3.org/2005/Atom">
							</feed>
						`)),
					},
				},
			},
			expectedFeeds: FeedResultMap{
				repo1.FeedURL: GitRepoResult{
					RepoName:          repo1.Name,
					RelFeedHasEntries: true,
				},
				repo2.FeedURL: GitRepoResult{
					RepoName: repo2.Name,
				},
			},
		},
		{
			name: "404 response should trigger an error",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred`,
					Response: http.Response{
						Body:       io.NopCloser(strings.NewReader(``)),
						Status:     testutils.StatusNotFoundString,
						StatusCode: http.StatusNotFound,
					},
				},
			},
			expectedFeeds: nil,
		},
		{
			name: "Reading response body should trigger an error",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred`,
					Response: http.Response{
						Body:       testutils.NewErrorReadCloser(),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
			},
			expectedFeeds: nil,
		},
		{
			name: "Invalid json should trigger an error",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred`,
					Response: http.Response{
						Body:       io.NopCloser(strings.NewReader(testutils.Invalid)),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
			},
			expectedFeeds: nil,
		},
		{
			name: "Pagination across multiple pages",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred\?per_page=100$`,
					MaxMatches: 1,
					Response: http.Response{
						Header: http.Header{
							"Link": []string{
								`<https://api.github.com/user/starred?page=2>; rel="next"`,
							},
						},
						Body: io.NopCloser(strings.NewReader(`[
							{
								"name": "` + repo1.Name.String() + `",
								"html_url": "` + repo1.RepoURL.String() + `"
							}
						]`)),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
				{
					UrlPattern: `page=2`,
					Response: http.Response{
						Body: io.NopCloser(strings.NewReader(`[
							{
								"name": "` + repo2.Name.String() + `",
								"html_url": "` + repo2.RepoURL.String() + `"
							}
						]`)),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
				{
					UrlPattern: repo1.Name.String() + `/releases\.atom`,
					Response: http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`
							<feed xmlns="http://www.w3.org/2005/Atom">
								<entry>
									<title>Release 1</title>
									<id>1</id>
								</entry>
							</feed>
						`)),
					},
				},
				{
					UrlPattern: repo2.Name.String() + `/releases\.atom`,
					Response: http.Response{
						StatusCode: http.StatusOK,
						Body: io.NopCloser(strings.NewReader(`
							<feed xmlns="http://www.w3.org/2005/Atom">
								<entry>
									<title>Release 2</title>
									<id>2</id>
								</entry>
							</feed>
						`)),
					},
				},
			},
			expectedFeeds: FeedResultMap{
				repo1.FeedURL: GitRepoResult{
					RepoName:          repo1.Name,
					RelFeedHasEntries: true,
				},
				repo2.FeedURL: GitRepoResult{
					RepoName:          repo2.Name,
					RelFeedHasEntries: true,
				},
			},
		},
		{
			name: "Repo with failing release feed",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: `api\.github\.com/user/starred`,
					Response: http.Response{
						Body: io.NopCloser(strings.NewReader(`[
							{
								"name": "` + repo1.Name.String() + `",
								"html_url": "` + repo1.RepoURL.String() + `"
							}
						]`)),
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
				{
					UrlPattern: repo1.Name.String() + `/releases\.atom`,
					Response: http.Response{
						StatusCode: http.StatusNotFound,
						Status:     testutils.StatusNotFoundString,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
			},
			expectedFeeds: FeedResultMap{
				repo1.FeedURL: GitRepoResult{
					RepoName: repo1.Name,
					Err:      common.HTTPError{StatusCode: http.StatusNotFound},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			mockTransport := testutils.NewMockRoutedResponseRoundTripper(tc.mocks)
			mockClient := &http.Client{Transport: &mockTransport}

			gh := NewGitForgeClient(
				GitHubForgeType,
				testutils.GitHubFqdn,
				testutils.GitHubToken,
				testutils.TestLogger(t),
				mockClient,
			)

			actual, err := gh.LoadFeeds(ctx)

			if tc.expectedFeeds == nil {
				if err == nil {
					t.Fatalf("Expected an error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(tc.expectedFeeds) != len(actual) {
				t.Errorf("Expected %d results, got %d", len(tc.expectedFeeds), len(actual))
				return
			}

			for feedURL, expectedResult := range tc.expectedFeeds {
				actualResult, exists := actual[feedURL]
				if !exists {
					t.Errorf("Expected feed %s not found in results", feedURL)
					continue
				}

				if !expectedResult.Equal(actualResult) {
					t.Errorf(
						"Feed %s: expected %+v, got %+v",
						feedURL,
						expectedResult,
						actualResult,
					)
				}
			}
		})
	}
}
