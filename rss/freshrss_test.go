package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/testutils"
)

const (
	mockAuthToken = "1234567890"
	mockSid       = "2345678901"
)

func TestAuthenticate(t *testing.T) {
	testCases := []struct {
		name        string
		responses   []http.Response
		expectError bool
	}{
		{
			name: "Successful authentication",
			responses: []http.Response{
				{
					Body: io.NopCloser(
						strings.NewReader(fmt.Sprintf("Auth=%s\nSID=%s\n", mockAuthToken, mockSid)),
					),
					StatusCode: http.StatusOK,
					Status:     testutils.StatusOKString,
				},
			},
			expectError: false,
		},
		{
			name: "Invalid text response should return error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("Invalid response")),
					StatusCode: http.StatusOK,
					Status:     testutils.StatusOKString,
				},
			},
			expectError: true,
		},
		{
			name: "Failed authentication",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("Error=BadAuthentication\n")),
					StatusCode: http.StatusUnauthorized,
					Status:     testutils.StatusUnauthorizedString,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.responses[0].Status, func(t *testing.T) {
			t.Parallel()
			mockTransport := testutils.NewMockMultiResponseRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}

			f := NewFreshRSSClient(
				testutils.FreshRSSUser,
				testutils.FreshRSSURL,
				testutils.TestLogger(t),
				mockClient,
			)
			err := f.Authenticate(context.Background(), testutils.FreshRSSToken)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error but got %v", err)
				return
			}

		})
	}
}

func TestAddFeed(t *testing.T) {

	testCases := []struct {
		name        string
		mocks       []testutils.MockRoutedResponse
		expectError bool
	}{
		{
			name: "Successful feed addition",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: ".*quickadd",
					Response: http.Response{
						Body: io.NopCloser(strings.NewReader(`
					{
						"query": "http://localhost/feeds/123",
						"numResults": 1,
						"streamId": "feed/http://localhost/feeds/123",
						"streamName": "name"
					}
					`)),
						StatusCode: http.StatusOK,
						Status:     testutils.StatusOKString,
					},
				},
				{
					UrlPattern: ".*edit",
					Response: http.Response{
						Status:     testutils.StatusOKString,
						StatusCode: http.StatusOK,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Failed feed addition on step 1",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: ".*quickadd",
					Response: http.Response{
						Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
						StatusCode: http.StatusUnauthorized,
						Status:     testutils.StatusUnauthorizedString,
					},
				},
			},
			expectError: true,
		},
		{
			name: "Failed feed addition on step 2",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: ".*quickadd",
					Response: http.Response{
						Body: io.NopCloser(strings.NewReader(`
					{
						"query": "http://localhost/feeds/123",
						"numResults": 1,
						"streamId": "feed/http://localhost/feeds/123",
						"streamName": "name"
					}`)),
						StatusCode: http.StatusOK,
						Status:     testutils.StatusOKString,
					},
				},
				{
					UrlPattern: ".*edit",
					Response: http.Response{
						Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
						StatusCode: http.StatusBadRequest,
						Status:     "400 Bad Request",
					},
				},
			},
			expectError: true,
		},
		{
			name: "Failed feed with invalid response",
			mocks: []testutils.MockRoutedResponse{
				{
					UrlPattern: ".*quickadd",
					Response: http.Response{
						Body:       io.NopCloser(strings.NewReader(`Invalid response`)),
						StatusCode: http.StatusOK,
						Status:     testutils.StatusOKString,
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockRoutedResponseRoundTripper(tc.mocks)
			mockClient := &http.Client{Transport: &mockTransport}

			f := NewFreshRSSClient(
				testutils.FreshRSSUser,
				testutils.FreshRSSURL,
				testutils.TestLogger(t),
				mockClient,
			)

			err := f.AddFeed(ctx, "http://localhost/feeds/123", "name", "category")
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

func TestLoadFeeds(t *testing.T) {

	testCases := []struct {
		name            string
		gitForge        string
		responses       []http.Response
		expectedFeedMap *common.Set[common.FeedURL]
		expectError     bool
	}{
		{
			name:     "Successful feed retrieval",
			gitForge: "GitHub",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`
						{
							"subscriptions": [
								{
									"url": "http://localhost/feeds/123",
									"categories": [
										{
											"label": "GitHub"
										}
									]
								},
								{
									"url": "http://localhost/feeds/456",
									"categories": [
										{
											"label": "GitHub"
										}
									]
								}
							]
						}`),
					),
					StatusCode: http.StatusOK,
					Status:     testutils.StatusOKString,
				},
			},
			expectedFeedMap: common.NewSet[common.FeedURL](
				"http://localhost/feeds/123",
				"http://localhost/feeds/456",
			),
			expectError: false,
		},
		{
			name: "Failed feed retrieval",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"error": "error"}`)),
					StatusCode: http.StatusUnauthorized,
					Status:     testutils.StatusUnauthorizedString,
				},
			},
			expectedFeedMap: common.NewSet[common.FeedURL](),
			expectError:     true,
		},
		{
			name: "Failed feed with invalid response",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`Invalid response`)),
					StatusCode: http.StatusOK,
					Status:     testutils.StatusOKString,
				},
			},
			expectedFeedMap: common.NewSet[common.FeedURL](),
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockMultiResponseRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}

			f := NewFreshRSSClient(
				testutils.FreshRSSUser,
				testutils.FreshRSSURL,
				testutils.TestLogger(t),
				mockClient,
			)

			feeds, err := f.LoadFeeds(ctx, FeedCategory(tc.gitForge))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}

				if feeds.Len() != tc.expectedFeedMap.Len() {
					t.Errorf("Expected %d feeds but got %d", tc.expectedFeedMap.Len(), feeds.Len())
				}

				for feed := range feeds.All() {
					if !tc.expectedFeedMap.Contains(feed) {
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
		feedURL     common.FeedURL
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
					Status:     testutils.StatusOKString,
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
					Status:     testutils.StatusUnauthorizedString,
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mockTransport := testutils.NewMockMultiResponseRoundTripper(tc.responses)
			mockClient := &http.Client{Transport: &mockTransport}

			f := NewFreshRSSClient(
				testutils.FreshRSSUser,
				testutils.FreshRSSURL,
				testutils.TestLogger(t),
				mockClient,
			)

			err := f.RemoveFeed(ctx, tc.feedURL)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error but got %v", err)
			}
		})
	}
}
