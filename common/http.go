package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// This is a common HTTP method that can be used by any of our client objects.
func DoAPIRequest(
	ctx context.Context,
	method string,
	reqURL string,
	payload []byte,
	headers http.Header,
	client *http.Client,
) ([]byte, http.Header, error) {
	var req *http.Request
	var err error
	if payload != nil {
		reader := bytes.NewReader(payload)
		req, err = http.NewRequestWithContext(ctx, method, reqURL, reader)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, reqURL, nil)
	}
	if err != nil {
		return nil, nil, err
	}

	// Set headers
	req.Header = headers.Clone()

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close() // nolint: errcheck

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	// If the status code is OK we can return no error
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return data, res.Header, nil
	}

	return data, res.Header, HTTPError{
		URL:        reqURL,
		StatusCode: res.StatusCode,
		Status:     res.Status,
	}
}

// Sometimes we need to know if the error is 404 or something else...
type HTTPError struct {
	URL        string
	StatusCode int
	Status     string
}

func (h HTTPError) Error() string {
	return fmt.Sprintf(
		"http error, url: %s, status code: %d, status: %s",
		h.URL, h.StatusCode, h.Status,
	)
}
