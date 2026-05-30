package githost

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func DoAPIRequest(
	ctx context.Context,
	reqURL string,
	headers map[string]string,
	nextPageLinkRegex *regexp.Regexp,
	client *http.Client,
) (*GitHostResponse, error) {
	// No request will always be valid here so we can ignore the error
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("git host returned an http error code %d", res.StatusCode)
	}

	return processResponse(res, nextPageLinkRegex)
}

func processResponse(
	r *http.Response,
	nextPageLinkRegex *regexp.Regexp,
) (*GitHostResponse, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	linkRaw := r.Header.Get("Link")
	links := strings.SplitSeq(linkRaw, ",")
	for link := range links {
		matches := nextPageLinkRegex.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GitHostResponse{Data: data, NextPage: matches[1]}, nil
		}
	}

	return &GitHostResponse{Data: data}, nil
}
