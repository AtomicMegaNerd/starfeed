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
		},
		{
			name:    "Feed has no entries",
			feedUrl: "http://example.com/feed",
			responses: []http.Response{
				{
					Body: io.NopCloser(strings.NewReader(`
						<feed xmlns="http://www.w3.org/2005/Atom">
						</feed>
					`)),
				},
			},
			expectHasEntries: false,
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
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fc := tc.GetTestObject()
			hasEntries := fc.CheckFeedHasEntries(tc.feedUrl)

			if hasEntries != tc.expectHasEntries {
				t.Errorf("Expected %t, got %t", tc.expectHasEntries, hasEntries)
			}
		})
	}
}
