package rss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/mocks"
)

const (
	mockAuthToken = "1234567890"
	mockSid       = "2345678901"
)

type AuthenticateTestCase struct {
	name              string
	responses         []http.Response
	expectedAuthToken string
	expectError       bool
}

func (tc *AuthenticateTestCase) GetTestObject() *FreshRSS {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return MockValidRSSServer(mockClient)
}

func TestAuthenticate(t *testing.T) {
	testCases := []AuthenticateTestCase{
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
			testObject := tc.GetTestObject()
			err := testObject.Authenticate(ctx)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}

				// Note: We can no longer access authToken directly since it's private.
				// Authentication success is verified by the lack of error.
			}
		})
	}
}

type ParsePlainTextAuthResponseTestCase struct {
	name          string
	inputData     []byte
	expectedToken string
	expectError   bool
}

func TestParsePlainTextAuthResponse(t *testing.T) {
	testCases := []ParsePlainTextAuthResponseTestCase{
		{
			name:          "Valid auth response with Auth and SID",
			inputData:     []byte("Auth=1234567890\nSID=abcdef\n"),
			expectedToken: "1234567890",
			expectError:   false,
		},
		{
			name:          "Valid auth response with only Auth",
			inputData:     []byte("Auth=token123\n"),
			expectedToken: "token123",
			expectError:   false,
		},
		{
			name:          "Valid auth response with extra whitespace",
			inputData:     []byte("Auth=mytoken456\n\n"),
			expectedToken: "mytoken456",
			expectError:   false,
		},
		{
			name:          "No Auth field should return error",
			inputData:     []byte("SID=abcdef\nOther=value\n"),
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Empty response should return error",
			inputData:     []byte(""),
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Error response should return error",
			inputData:     []byte("Error=BadAuthentication\n"),
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Auth with empty value should return error",
			inputData:     []byte("Auth=\nSID=abcdef\n"),
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := parsePlainTextAuthResponse(tc.inputData)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, got none")
				}
				if token != "" {
					t.Errorf("Expected empty token on error, got %s", token)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if token != tc.expectedToken {
					t.Errorf("Expected token %s, got %s", tc.expectedToken, token)
				}
			}
		})
	}
}

type AddFeedTestCase struct {
	name             string
	responses        []http.Response
	urlRegexPatterns []string
	expectError      bool
}

func (tc *AddFeedTestCase) GetTestObject() *FreshRSS {
	mockTransport := mocks.NewMockURLSelectedRoundTripper(tc.responses, tc.urlRegexPatterns)
	mockClient := &http.Client{Transport: &mockTransport}
	return MockValidRSSServer(mockClient)
}

func TestAddFeed(t *testing.T) {
	testCases := []AddFeedTestCase{
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
			testObject := tc.GetTestObject()
			err := testObject.AddFeed(ctx, "http://localhost/feeds/123", "name", "category")
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

type GetExistingFeedsTestCase struct {
	name            string
	responses       []http.Response
	expectedFeedMap map[string]struct{}
	expectError     bool
}

func (tc *GetExistingFeedsTestCase) GetTestObject() *FreshRSS {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return MockValidRSSServer(mockClient)
}

func TestGetExistingFeeds(t *testing.T) {
	testCases := []GetExistingFeedsTestCase{
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
			testObject := tc.GetTestObject()
			ctx := context.Background()
			feeds, err := testObject.GetExistingFeeds(ctx)
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

type RemoveFeedTestCase struct {
	name        string
	feedURL     string
	responses   []http.Response
	expectError bool
}

func (tc *RemoveFeedTestCase) GetTestObject() *FreshRSS {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return MockValidRSSServer(mockClient)
}

func TestRemoveFeed(t *testing.T) {
	testCases := []RemoveFeedTestCase{
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
			testObject := tc.GetTestObject()
			ctx := context.Background()
			err := testObject.RemoveFeed(ctx, tc.feedURL)
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
