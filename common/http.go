package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

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

	switch res.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusAccepted:
		fallthrough
	case http.StatusCreated:
		return data, res.Header, nil
	}

	return data, res.Header, HTTPError{
		URL:        reqURL,
		Body:       string(data),
		StatusCode: res.StatusCode,
	}
}

type HTTPError struct {
	URL        string
	Body       string
	StatusCode int
}

func (h HTTPError) Error() string {
	return fmt.Sprintf(
		"http error, url: %s, body: %s, status code: %d",
		h.URL, h.Body, h.StatusCode,
	)
}
