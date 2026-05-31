package common

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/atomicmeganerd/starfeed/mocks"
)

func TestDoAPIRequest(t *testing.T) {
	testCases := []struct {
		name        string
		method      string
		reqURL      string
		payload     []byte
		headers     http.Header
		statusCode  int
		body        io.ReadCloser
		respHeaders http.Header
		expectError bool
	}{
		{
			name:        "GET 200 OK with no payload",
			method:      http.MethodGet,
			reqURL:      mocks.GitHubAPIURL + "/repos",
			payload:     nil,
			headers:     http.Header{"Authorization": {"token " + mocks.GitHubToken}},
			statusCode:  http.StatusOK,
			body:        io.NopCloser(strings.NewReader(`{"ok":true}`)),
			respHeaders: http.Header{"Content-Type": {"application/json"}},
		},
		{
			name:        "POST 201 Created with payload",
			method:      http.MethodPost,
			reqURL:      mocks.GitHubAPIURL + "/repos",
			payload:     []byte(`{"name":"test"}`),
			headers:     http.Header{"Content-Type": {"application/json"}},
			statusCode:  http.StatusCreated,
			body:        io.NopCloser(strings.NewReader(`{"id":1}`)),
			respHeaders: http.Header{},
		},
		{
			name:        "PUT 202 Accepted",
			method:      http.MethodPut,
			reqURL:      mocks.GitHubAPIURL + "/repos/1",
			payload:     []byte(`{"name":"updated"}`),
			headers:     http.Header{},
			statusCode:  http.StatusAccepted,
			body:        io.NopCloser(strings.NewReader("accepted")),
			respHeaders: http.Header{},
		},
		{
			name:        "404 Not Found returns error",
			method:      http.MethodGet,
			reqURL:      mocks.GitHubAPIURL + "/missing",
			payload:     nil,
			headers:     http.Header{},
			statusCode:  http.StatusNotFound,
			body:        io.NopCloser(strings.NewReader("not found")),
			respHeaders: http.Header{},
			expectError: true,
		},
		{
			name:        "500 Internal Server Error returns error",
			method:      http.MethodGet,
			reqURL:      mocks.GitHubAPIURL + "/broken",
			payload:     nil,
			headers:     http.Header{},
			statusCode:  http.StatusInternalServerError,
			body:        io.NopCloser(strings.NewReader("server error")),
			respHeaders: http.Header{},
			expectError: true,
		},
		{
			name:        "401 Unauthorized returns error",
			method:      http.MethodGet,
			reqURL:      mocks.GitHubAPIURL + "/private",
			payload:     nil,
			headers:     http.Header{},
			statusCode:  http.StatusUnauthorized,
			body:        io.NopCloser(strings.NewReader("unauthorized")),
			respHeaders: http.Header{},
			expectError: true,
		},
		{
			name:        "invalid URL returns error",
			method:      http.MethodGet,
			reqURL:      "://bad-url",
			payload:     nil,
			headers:     http.Header{},
			statusCode:  http.StatusOK,
			body:        io.NopCloser(strings.NewReader("")),
			respHeaders: http.Header{},
			expectError: true,
		},
		{
			name:        "body read error returns error",
			method:      http.MethodGet,
			reqURL:      mocks.GitHubAPIURL + "/repos",
			payload:     nil,
			headers:     http.Header{},
			statusCode:  http.StatusOK,
			body:        mocks.NewErrorReadCloser(),
			respHeaders: http.Header{},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var readBuf bytes.Buffer
			responses := []http.Response{
				{
					StatusCode: tc.statusCode,
					Body:       io.NopCloser(io.TeeReader(tc.body, &readBuf)),
					Header:     tc.respHeaders,
				},
			}
			mockTransport := mocks.NewMockRoundTripper(responses)
			client := &http.Client{Transport: &mockTransport}

			body, respHeaders, err := DoAPIRequest(
				context.Background(), tc.method, tc.reqURL, tc.payload, tc.headers, client,
			)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected an error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if string(body) != readBuf.String() {
				t.Errorf("expected body %q, got %q", readBuf.String(), string(body))
			}

			assertHeaders(t, tc.respHeaders, respHeaders)
		})
	}
}

func assertHeaders(t *testing.T, want, got http.Header) {
	t.Helper()
	for key, wantVals := range want {
		gotVals := got.Values(key)
		if !slices.Equal(gotVals, wantVals) {
			t.Errorf("header %q: want %v, got %v", key, wantVals, gotVals)
		}
	}
}
