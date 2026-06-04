package githost

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/mocks"
	"github.com/lmittmann/tint"
)

func TestCheckReleaseFeed(t *testing.T) {
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
			feedURL: "",
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
					Body: mocks.NewErrorReadCloser(),
				},
			},
			expectError: true,
		},
		{
			name:    "Error parsing XML",
			repoURL: "https://github.com/user/repo5",
			feedURL: "https://github.com/user/repo6/releases.atom",
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
			gh := MockValidGitHub(mockClient, logger)

			repo := StarredRepo{RepoURL: tc.repoURL}
			err := gh.CheckReleaseFeed(ctx, &repo)

			if err == nil {
				if tc.expectError {
					t.Fatalf("Expected an error, got none")
				}
			}

			if err != nil && !tc.expectError {
				t.Fatalf("Expected no error, got %v", err)
			}
		})
	}
}
