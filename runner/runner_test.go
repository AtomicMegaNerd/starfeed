package runner

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/github"
	"github.com/atomicmeganerd/starfeed/mocks"
)

const (
	mockGhToken       = "gh_token"
	mockFreshRssUrl   = "http://freshrss.example.com"
	mockFreshRssUser  = "testuser"
	mockFreshRssToken = "freshrss_token"
)

type QueryAndPublishFeedsTestCase struct {
	name      string
	responses []http.Response
	urlRegex  []string
}

func (tc *QueryAndPublishFeedsTestCase) GetTestObject() RepoRSSPublisher {
	mockTransport := mocks.NewMockUrlSelectedRoundTripper(tc.responses, tc.urlRegex)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewRepoRSSPublisher(
		mockGhToken,
		mockFreshRssUrl,
		mockFreshRssUser,
		mockFreshRssToken,
		context.Background(),
		mockClient,
	)
}

func TestQueryAndPublishFeeds(t *testing.T) {
	testCases := []QueryAndPublishFeedsTestCase{
		{
			name: "Successful workflow with no repos",
			responses: []http.Response{
				// FreshRSS auth request
				{
					Body:       io.NopCloser(strings.NewReader(`{"auth_token": "test_token"}`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
				// FreshRSS get existing feeds
				{
					Body:       io.NopCloser(strings.NewReader(`{"feeds": []}`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
				// GitHub get starred repos
				{
					Body:       io.NopCloser(strings.NewReader(`[]`)),
					Status:     "200 OK",
					StatusCode: http.StatusOK,
				},
			},
			urlRegex: []string{
				`.*freshrss.*api.*auth.*`,  // FreshRSS auth
				`.*freshrss.*api.*feeds.*`, // FreshRSS feeds
				`.*api\.github\.com.*`,     // GitHub API
			},
		},
		{
			name: "Authentication failure should exit early",
			responses: []http.Response{
				// FreshRSS auth request fails
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "unauthorized"}`)),
					Status:     "401 Unauthorized",
					StatusCode: http.StatusUnauthorized,
				},
			},
			urlRegex: []string{
				`.*freshrss.*api.*auth.*`, // FreshRSS auth
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisher := tc.GetTestObject()

			// This should not panic or hang
			publisher.QueryAndPublishFeeds()
		})
	}
}

func TestNewRepoRSSPublisher(t *testing.T) {
	mockClient := &http.Client{}
	ctx := context.Background()

	publisher := NewRepoRSSPublisher(
		mockGhToken,
		mockFreshRssUrl,
		mockFreshRssUser,
		mockFreshRssToken,
		ctx,
		mockClient,
	)

	if publisher.ghToken != mockGhToken {
		t.Errorf("Expected ghToken %s, got %s", mockGhToken, publisher.ghToken)
	}

	if publisher.freshRssUrl != mockFreshRssUrl {
		t.Errorf("Expected freshRssUrl %s, got %s", mockFreshRssUrl, publisher.freshRssUrl)
	}

	if publisher.freshRssUser != mockFreshRssUser {
		t.Errorf("Expected freshRssUser %s, got %s", mockFreshRssUser, publisher.freshRssUser)
	}

	if publisher.freshRssToken != mockFreshRssToken {
		t.Errorf("Expected freshRssToken %s, got %s", mockFreshRssToken, publisher.freshRssToken)
	}

	if publisher.ctx != ctx {
		t.Error("Expected context to match")
	}

	if publisher.client != mockClient {
		t.Error("Expected client to match")
	}
}

type FilterOutNonGithubFeedsTestCase struct {
	name           string
	inputFeeds     map[string]struct{}
	expectedFeeds  map[string]struct{}
	expectedLength int
}

func TestFilterOutNonGithubFeeds(t *testing.T) {
	testCases := []FilterOutNonGithubFeedsTestCase{
		{
			name: "All feeds are GitHub releases",
			inputFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://github.com/user/repo3/releases.atom": {},
			},
			expectedLength: 3,
		},
		{
			name: "Mixed GitHub and non-GitHub feeds",
			inputFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://example.com/feed.xml":                {},
				"https://github.com/user/repo2/releases.atom": {},
				"https://blog.example.com/rss":                {},
			},
			expectedFeeds: map[string]struct{}{
				"https://github.com/user/repo1/releases.atom": {},
				"https://github.com/user/repo2/releases.atom": {},
			},
			expectedLength: 2,
		},
		{
			name: "No GitHub feeds",
			inputFeeds: map[string]struct{}{
				"https://example.com/feed.xml":  {},
				"https://blog.example.com/rss":  {},
				"https://news.example.com/atom": {},
			},
			expectedFeeds:  map[string]struct{}{},
			expectedLength: 0,
		},
		{
			name:           "Empty input map",
			inputFeeds:     map[string]struct{}{},
			expectedFeeds:  map[string]struct{}{},
			expectedLength: 0,
		},
		{
			name: "GitHub feeds with dots and dashes in names",
			inputFeeds: map[string]struct{}{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
				"https://example.com/feed.xml":                             {},
			},
			expectedFeeds: map[string]struct{}{
				"https://github.com/nix-community/NixOS-WSL/releases.atom": {},
				"https://github.com/EdenEast/nightfox.nvim/releases.atom":  {},
			},
			expectedLength: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create GitHub builder for the test
			mockClient := &http.Client{}
			gh := github.NewGitHubStarredFeedBuilder("token", context.Background(), mockClient)

			// Call the function under test
			result := filterOutNonGithubFeeds(gh, tc.inputFeeds)

			// Check the length
			if len(result) != tc.expectedLength {
				t.Errorf("Expected %d feeds, got %d", tc.expectedLength, len(result))
			}

			// Check each expected feed exists in result
			for expectedFeed := range tc.expectedFeeds {
				if _, exists := result[expectedFeed]; !exists {
					t.Errorf("Expected feed %s to be in result, but it wasn't", expectedFeed)
				}
			}

			// Check no unexpected feeds exist in result
			for resultFeed := range result {
				if _, exists := tc.expectedFeeds[resultFeed]; !exists {
					t.Errorf("Unexpected feed %s found in result", resultFeed)
				}
			}
		})
	}
}
