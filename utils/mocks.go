package mocks

import (
	"bytes"
	"io"
	"net/http"
)

type MockRoundTripper struct {
	response   string
	statusCode int
	headers    http.Header
}

func NewMockRoundTripper(response string, statusCode int, headers http.Header) MockRoundTripper {
	return MockRoundTripper{response, statusCode, headers}
}

func (mrt *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: mrt.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(mrt.response)),
		Header:     mrt.headers,
	}, nil
}
