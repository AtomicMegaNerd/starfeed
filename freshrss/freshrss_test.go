package freshrss

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
	mockBaseUrl    = "http://localhost"
	invalidBaseURL = "protocol+not_a_url"
	mockUser       = "user"
	mockApiToken   = "token"
	mockAuthToken  = "1234567890"
	mockSid        = "2345678901"
)

type AuthenticateTestCase struct {
	name               string
	responses          []http.Response
	expxectedAuthToken string
	expectError        bool
}

func (tc *AuthenticateTestCase) GetTestObject() *FreshRSSFeedManager {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewFreshRSSFeedManager(
		mockBaseUrl, mockUser, mockApiToken, context.Background(), mockClient,
	)
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
					Status:     "200 OK",
				},
			},
			expxectedAuthToken: mockAuthToken,
			expectError:        false,
		},
		{
			name: "Invalid text response should return error",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("Invalid response")),
					StatusCode: http.StatusOK,
					Status:     "200 OK",
				},
			},
			expxectedAuthToken: "",
			expectError:        true,
		},
		{
			name: "Failed authentication",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("Error=BadAuthentication\n")),
					StatusCode: http.StatusUnauthorized,
					Status:     "401 Unauthorized",
				},
			},
			expxectedAuthToken: "",
			expectError:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.responses[0].Status, func(t *testing.T) {
			testObject := tc.GetTestObject()
			err := testObject.Authenticate()
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got %v", err)
				}

				if testObject.authToken != tc.expxectedAuthToken {
					t.Errorf("Expected auth token %s but got %s", tc.expxectedAuthToken, testObject.authToken)
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

func (tc *AddFeedTestCase) GetTestObject() *FreshRSSFeedManager {
	mockTransport := mocks.NewMockUrlSelectedRoundTripper(tc.responses, tc.urlRegexPatterns)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewFreshRSSFeedManager(
		mockBaseUrl, mockUser, mockApiToken, context.Background(), mockClient,
	)
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
					Status:     "200 OK",
				},
				{
					Status:     "200 OK",
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
					Status:     "401 Unauthorized",
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
					Status:     "200 OK",
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
					Status:     "200 OK",
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
			testObject := tc.GetTestObject()
			err := testObject.AddFeed("http://localhost/feeds/123", "name", "category")
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

func (tc *GetExistingFeedsTestCase) GetTestObject() *FreshRSSFeedManager {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewFreshRSSFeedManager(
		mockBaseUrl, mockUser, mockApiToken, context.Background(), mockClient,
	)
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
					Status:     "200 OK",
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
					Status:     "401 Unauthorized",
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
					Status:     "200 OK",
				},
			},
			expectedFeedMap: map[string]struct{}{},
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testObject := tc.GetTestObject()
			feeds, err := testObject.GetExistingFeeds()
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
	feedUrl     string
	responses   []http.Response
	expectError bool
}

func (tc *RemoveFeedTestCase) GetTestObject() *FreshRSSFeedManager {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewFreshRSSFeedManager(
		mockBaseUrl, mockUser, mockApiToken, context.Background(), mockClient,
	)
}

func TestRemoveFeed(t *testing.T) {
	testCases := []RemoveFeedTestCase{
		{
			name:    "Successful feed removal",
			feedUrl: "http://localhost/feeds/124",
			responses: []http.Response{
				{
					Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
					StatusCode: http.StatusOK,
					Status:     "200 OK",
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
					Status:     "401 Unauthorized",
				},
			},
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testObject := tc.GetTestObject()
			err := testObject.RemoveFeed(tc.feedUrl)
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

type DoApiRequestTestCase struct {
	name        string
	url         string
	responses   []http.Response
	expectError bool
}

func (tc *DoApiRequestTestCase) GetTestObject() *FreshRSSFeedManager {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewFreshRSSFeedManager(
		mockBaseUrl, mockUser, mockApiToken, context.Background(), mockClient,
	)
}

func TestDoApiRequest(t *testing.T) {

	testCases := []DoApiRequestTestCase{
		{
			name: "Successful request",
			url:  fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", mockBaseUrl),
			responses: []http.Response{
				{
					Body: io.NopCloser(
						strings.NewReader(fmt.Sprintf("Auth=%s\nSID=%s\n", mockAuthToken, mockSid)),
					),
					StatusCode: http.StatusOK,
					Status:     "200 OK",
				},
			},
			expectError: false,
		},
		{
			name: "Error reading response body",
			url:  fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", mockBaseUrl),
			responses: []http.Response{
				{
					Body:       mocks.NewErrorReadCloser(),
					StatusCode: http.StatusOK,
					Status:     "200 OK",
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testObject := tc.GetTestObject()
			_, err := testObject.doApiRequest(tc.url, nil, false)
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
