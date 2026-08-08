package testutils

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
)

const (
	StatusOKString           = "200 OK"
	StatusNotFoundString     = "404 Not Found"
	StatusUnauthorizedString = "401 Unauthorized"
	StatusIServerErrorString = "500 Internal Server Error"
)

// This is a mock round tripper that can be used to mock http responses
// for testing purposes. It is used to mock the http.Client's transport
// in the http.Client.Do method. It can be used to mock multiple responses
// in a single test.
type MockMultiResponseRoundTripper struct {
	responses []http.Response
	calls     atomic.Int32
}

func NewMockMultiResponseRoundTripper(responses []http.Response) MockMultiResponseRoundTripper {
	return MockMultiResponseRoundTripper{responses: responses}
}

func (mrt *MockMultiResponseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Increment and store in the same op to prevent a race
	calls := int(mrt.calls.Add(1) - 1)
	if calls >= len(mrt.responses) {
		return nil, errors.New("no more responses in mock responses slice")
	}
	res := &mrt.responses[calls]
	return res, nil
}

func (mtr *MockMultiResponseRoundTripper) GetNumCalls() int {
	return int(mtr.calls.Load())
}

// This is a mock round tripper that can be used to mock http responses based on the URL
// of the request. We will use regex patterns to match the URL of the requests. We can set
// a max matches attribute as well which will specify how many times we can match on the same
// pattern
type MockRoutedResponse struct {
	Response   http.Response
	UrlPattern string
	Err        error

	// We increment this value so we want to be careful
	Matches    atomic.Int32
	MaxMatches int
}

type MockRoutedResponseRoundTripper struct {
	resps []MockRoutedResponse
}

func NewMockRoutedResponseRoundTripper(
	responses []MockRoutedResponse,
) MockRoutedResponseRoundTripper {
	return MockRoutedResponseRoundTripper{resps: responses}
}

func (t *MockRoutedResponseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for ix := range t.resps {
		resp := &t.resps[ix]
		matches, _ := regexp.MatchString(resp.UrlPattern, req.URL.String())
		if matches {
			// Increment and store in the same op to prevent a race.
			matches := int(resp.Matches.Add(1) - 1)
			if resp.MaxMatches == 0 || matches < resp.MaxMatches {
				return &resp.Response, resp.Err
			}
		}
	}
	// Return not found if we don't match which is what would happen
	return nil, fmt.Errorf("no response for url: %s", req.URL.String())
}

// This is a mock ReadCloser that can be used to mock an error when reading
// from the response body. It is used to test error handling when reading
// from the response body.
type ErrorReadCloser struct{}

func NewErrorReadCloser() ErrorReadCloser {
	return ErrorReadCloser{}
}

func (erc ErrorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("error reading from response body")
}

func (erc ErrorReadCloser) Close() error {
	return nil
}
