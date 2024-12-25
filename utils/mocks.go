package mocks

import (
	"errors"
	"net/http"
	"sync"
)

// This is a mock round tripper that can be used to mock http responses
// for testing purposes. It is used to mock the http.Client's transport
// in the http.Client.Do method. It can be used to mock multiple responses
// in a single test.

type MockRoundTripper struct {
	responses []http.Response
	calls     int
	mtx       sync.Mutex
}

func NewMockRoundTripper(responses []http.Response) MockRoundTripper {
	return MockRoundTripper{responses: responses}
}

func (mrt *MockRoundTripper) Increment() {
	mrt.mtx.Lock()
	defer mrt.mtx.Unlock()
	mrt.calls += 1
}

func (mrt *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if mrt.calls >= len(mrt.responses) {
		return nil, errors.New("no more responses in mock responses slice")
	}
	defer mrt.Increment()
	return &mrt.responses[mrt.calls], nil
}
