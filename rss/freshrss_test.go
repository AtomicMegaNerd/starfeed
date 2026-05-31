package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/mocks"
)

const (
	mockAuthToken = "1234567890"
	mockSid       = "2345678901"
)

func TestAuthenticate(t *testing.T) {
	testCases := []struct {
		name              string
		responses         []http.Response
		expectedAuthToken string
		expectError       bool
	}{
		{
			name: "Successful authentication",
			responses: []http.Response{
				{
					Body: io.NopCloser(
						strings.NewReader(fmt.Sprintf("Auth=%s\nSID=%s\n", mockAuthToken, mockSid)),
					),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
			},
			expectedAuthToken: mockAuthToken,
			expectError:       false,
		},
		{
			name: "Invalid text response should return error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("Invalid response")),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
			},
			expectedAuthToken: "",
			expectError:       true,
		},
		{
			name: "Failed authentication",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("Error=BadAuthentication\n")),
					StatusCode: http.StatusUnauthorized,
					Status:     mocks.StatusUnauthorizedString,
				},
			},
			expectedAuthToken: "",
			expectError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.responses[0].Status, func(t *testing.T) {
			ctx := context.Background()
			mockTransport := mocks.NewMockRoundTripper(tc.responses)

			headers := http.Header{}
			headers.Set("Content-type", "application/x-www-form-urlencoded")
			authToken, err := authenticate(
				ctx,
				config.MockValidFreshRSSConfig,
				headers,
				mocks.TestLogger(),
				&http.Client{Transport: &mockTransport},
			)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}

				if authToken != tc.expectedAuthToken {
					t.Errorf("expected authToken %s but got %s", authToken, tc.expectedAuthToken)
				}
			}
		})
	}
}

func TestAddFeed(t *testing.T) {
	testCases := []struct {
		name             string
		responses        []http.Response
		urlRegexPatterns []string
		expectError      bool
	}{
		{
			name: "Successful feed addition",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`
					{
						"query": "http://localhost/feeds/123",
						"numResults": 1,
						"streamId": "feed/http://localhost/feeds/123",
						"streamName": "name"
					}
					`)),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
				{
					Status:     mocks.StatusOKString,
					StatusCode: http.StatusOK,
				},
			},
			urlRegexPatterns: []string{
				".*quickadd",
				".*edit",
			},
			expectError: false,
		},
		{
			name: "Failed feed addition on step 1",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
					StatusCode: http.StatusUnauthorized,
					Status:     mocks.StatusUnauthorizedString,
				},
			},
			urlRegexPatterns: []string{
				".*quickadd",
			},
			expectError: true,
		},
		{
			name: "Failed feed addition on step 2",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`
					{
						"query": "http://localhost/feeds/123",
						"numResults": 1,
						"streamId": "feed/http://localhost/feeds/123",
						"streamName": "name"
					}`)),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
				},
			},
			urlRegexPatterns: []string{
				".*quickadd",
				".*edit",
			},
			expectError: true,
		},
		{
			name: "Failed feed with invalid response",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`Invalid response`)),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
			},
			urlRegexPatterns: []string{
				".*quickadd",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockTransport := mocks.NewMockURLSelectedRoundTripper(tc.responses, tc.urlRegexPatterns)
			mockClient := &http.Client{Transport: &mockTransport}
			rss := MockValidRSSServer(mockClient)
			err := rss.AddFeed(ctx, "http://localhost/feeds/123", "name", "category")
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
			}
		})
	}
}

func TestGetExistingFeeds(t *testing.T) {
	testCases := []struct {
		name            string
		responses       []http.Response
		expectedFeedMap map[string]struct{}
		expectError     bool
	}{
		{
			name: "Successful feed retrieval",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`
						{
							"subscriptions": [
								{
									"url": "http://localhost/feeds/123"
								},
								{
									"url": "http://localhost/feeds/456"
								}
							]
						}`),
					),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
			},
			expectedFeedMap: map[string]struct{}{
				"http://localhost/feeds/123": {},
				"http://localhost/feeds/456": {},
			},
			expectError: false,
		},
		{
			name: "Failed feed retrieval",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
					StatusCode: http.StatusUnauthorized,
					Status:     mocks.StatusUnauthorizedString,
				},
			},
			expectedFeedMap: map[string]struct{}{},
			expectError:     true,
		},
		{
			name: "Failed feed with invalid response",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`Invalid response`)),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
			},
			expectedFeedMap: map[string]struct{}{},
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockTransport := mocks.NewMockRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}
			rss := MockValidRSSServer(mockClient)
			ctx := context.Background()
			feeds, err := rss.GetExistingFeeds(ctx)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}

				if len(feeds) != len(tc.expectedFeedMap) {
					t.Errorf("Expected %d feeds but got %d", len(tc.expectedFeedMap), len(feeds))
				}

				for feed := range feeds {
					if _, ok := tc.expectedFeedMap[feed]; !ok {
						t.Errorf("Unexpected feed %s", feed)
					}
				}
			}
		})
	}
}

func TestRemoveFeed(t *testing.T) {
	testCases := []struct {
		name        string
		feedURL     string
		responses   []http.Response
		expectError bool
	}{
		{
			name:    "Successful feed removal",
			feedURL: "http://localhost/feeds/124",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
					StatusCode: http.StatusOK,
					Status:     mocks.StatusOKString,
				},
			},
			expectError: false,
		},
		{
			name: "Failure response should return error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
					StatusCode: http.StatusUnauthorized,
					Status:     mocks.StatusUnauthorizedString,
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
			rss := MockValidRSSServer(mockClient)
			err := rss.RemoveFeed(ctx, tc.feedURL)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}
			}
		})
	}
}
