package githost

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/mocks"
)

func TestAddFeedURLToRepo(t *testing.T) {
	testCases := []struct {
		name             string
		feedURL          string
		responses        []http.Response
		expectHasEntries bool
		expectError      bool
	}{
		{
			name:    "Feed has entries",
			feedURL: "http://example.com/feed",
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
			feedURL: "http://example.com/feed",
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
			feedURL: "http://example.com/feed",
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
			feedURL: "http://example.com/feed",
			responses: []http.Response{
				{
					Body: mocks.NewErrorReadCloser(),
				},
			},
			expectError: true,
		},
		{
			name:    "Error parsing XML",
			feedURL: "http://example.com/feed",
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockTransport := mocks.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}
			gh := MockValidGitHub(mockClient)

			repo := StarredRepo{}

			err := gh.addReleaseFeedToRepo(ctx, &repo)

			if err != nil && !tc.expectError {
				t.Fatalf("Expected no error, got %v", err)
			}
			if err == nil && tc.expectError {
				t.Fatalf("Expected an error, got none")
			}

			if repo.FeedURL != tc.feedURL {
				t.Fatalf("Expected feedURL %s, got %s", tc.feedURL, repo.FeedURL)
			}
		})
	}
}
