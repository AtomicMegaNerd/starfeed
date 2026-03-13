package atom

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/mocks"
)

type CheckFeedHasEntriesTestCase struct {
	name             string
	feedUrl          string
	responses        []http.Response
	expectHasEntries bool
	expectError      bool
}

func (tc *CheckFeedHasEntriesTestCase) GetTestObject() AtomFeedChecker {
	mockTransport := mocks.NewMockRoundTripper(tc.responses)
	mockClient := &http.Client{Transport: &mockTransport}
	return NewAtomFeedChecker(context.Background(), mockClient)
}

func TestCheckFeedHasEntries(t *testing.T) {
	testCases := []CheckFeedHasEntriesTestCase{
		{
			name:    "Feed has entries",
			feedUrl: "http://example.com/feed",
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
			expectHasEntries: true,
			expectError:      false,
		},
		{
			name:    "Feed has no entries",
			feedUrl: "http://example.com/feed",
			responses: []http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(strings.NewReader(`
						<feed xmlns="http://www.w3.org/2005/Atom">
						</feed>
					`)),
				},
			},
			expectHasEntries: false,
			expectError:      false,
		},
		{
			name:    "Error making request",
			feedUrl: "http://example.com/feed",
			responses: []http.Response{
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(strings.NewReader("")),
				},
			},
			expectHasEntries: false,
			expectError:      true,
		},
		{
			name:    "Error reading response",
			feedUrl: "http://example.com/feed",
			responses: []http.Response{
				{
					Body: mocks.NewErrorReadCloser(),
				},
			},
			expectHasEntries: false,
			expectError:      true,
		},
		{
			name:    "Error parsing XML",
			feedUrl: "http://example.com/feed",
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
			expectHasEntries: false,
			expectError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fc := tc.GetTestObject()
			hasEntries, err := fc.CheckFeedHasEntries(tc.feedUrl)

			if err != nil && !tc.expectError {
				t.Fatalf("Expected no error, got %v", err)
			}
			if err == nil && tc.expectError {
				t.Fatalf("Expected an error, got none")
			}

			if hasEntries != tc.expectHasEntries {
				t.Errorf("Expected %t, got %t", tc.expectHasEntries, hasEntries)
			}
		})
	}
}
