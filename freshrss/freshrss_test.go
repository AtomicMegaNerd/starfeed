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
	mockBaseUrl   = "http://localhost"
	mockUser      = "user"
	mockApiToken  = "token"
	mockAuthToken = "1234567890"
	mockSid       = "2345678901"
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
