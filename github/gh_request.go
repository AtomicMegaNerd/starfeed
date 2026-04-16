package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func doApiRequest(
	ctx context.Context,
	client *http.Client,
	url string,
	token string,
	nextPageLinkRegex *regexp.Regexp,
) (*GitHubResponse, error) {
	headers := map[string]string{
		"Authorization":        fmt.Sprintf("Bearer %s", token),
		"X-GitHub-Api-Version": "2022-11-28",
		"User-Agent":           "github.com/atomicmeganerd/starfeed",
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}

	// No request will always be valid here so we can ignore the error
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("github returned an http error code %d", res.StatusCode)
	}

	ghResponse, err := processGitHubResponse(res, nextPageLinkRegex)
	if err != nil {
		return nil, err
	}

	return ghResponse, nil
}

func processGitHubResponse(
	r *http.Response,
	nextPageLinkRegex *regexp.Regexp,
) (*GitHubResponse, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	linkRaw := r.Header.Get("Link")
	links := strings.SplitSeq(linkRaw, ",")
	for link := range links {
		matches := nextPageLinkRegex.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GitHubResponse{data: data, nextPage: matches[1]}, nil
		}
	}

	return &GitHubResponse{data: data}, nil
}
