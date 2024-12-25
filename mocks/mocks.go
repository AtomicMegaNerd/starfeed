package mocks

import (
	"errors"
	"net/http"
	"regexp"
	"sync"
)

// This is a mock round tripper that can be used to mock http responses
// for testing purposes. It is used to mock the http.Client's transport
// in the http.Client.Do method. It can be used to mock multiple responses
// in a single test.
type MockMultiResponseRoundTripper struct {
	responses []http.Response
	calls     int
	mtx       sync.Mutex
}

func NewMockRoundTripper(responses []http.Response) MockMultiResponseRoundTripper {
	return MockMultiResponseRoundTripper{responses: responses}
}

func (mrt *MockMultiResponseRoundTripper) Increment() {
	mrt.mtx.Lock()
	defer mrt.mtx.Unlock()
	mrt.calls += 1
}

func (mrt *MockMultiResponseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if mrt.calls >= len(mrt.responses) {
		return nil, errors.New("no more responses in mock responses slice")
	}
	defer mrt.Increment()
	return &mrt.responses[mrt.calls], nil
}

// This is a mock round tripper that can be used to mock http responses based on the URL
// of the request. We will use regex patterns to match the URL of the requests.
type MockUrlSelectedRoundTripper struct {
	response         []http.Response
	urlRegexPatterns []string
}

func NewMockUrlSelectedRoundTripper(
	responses []http.Response, urls []string,
) MockUrlSelectedRoundTripper {
	return MockUrlSelectedRoundTripper{responses, urls}
}

func (ust *MockUrlSelectedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for ix, url := range ust.urlRegexPatterns {
		if matches, _ := regexp.MatchString(url, req.URL.String()); matches {
			return &ust.response[ix], nil
		}
	}
	return nil, errors.New("no response found for url")
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
