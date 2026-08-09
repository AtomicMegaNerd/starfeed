package gitforge

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/atomicmeganerd/starfeed/common"
)

// This regex will match if there is a next page in the response headers
var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

// GitForge struct represents a GitForge. We can load RSS feeds for all starred repos that
// belong to this Git Forge.
type GitForge struct {
	Name         ForgeName
	fetchRepoURL string
	headers      http.Header
	client       *http.Client
}

func NewGitForge(
	name ForgeName,
	forgeType ForgeType,
	fqdn, token string,
	client *http.Client,
) GitForge {
	return GitForge{
		Name:         name,
		fetchRepoURL: buildStarredRepoUrl(forgeType, fqdn),
		headers:      buildHeaders(forgeType, token),
		client:       client,
	}
}

// This will send the list of all feeds to our respChan
func (c GitForge) load(
	ctx context.Context,
) ([]Repo, error) {
	nextPageURL := c.fetchRepoURL
	allRepos := make([]Repo, 0)
	for {
		data, respHeaders, err := common.DoAPIRequest(
			ctx,
			http.MethodGet,
			nextPageURL,
			nil,
			c.headers,
			c.client,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"error %w getting raw data from gitforge url: %s", err, nextPageURL,
			)
		}

		repos := make([]Repo, 0)
		if err := json.Unmarshal(data, &repos); err != nil {
			return nil, fmt.Errorf(
				"error %w parsing JSON response from gitforge", err,
			)
		}

		allRepos = slices.Concat(allRepos, repos)

		nextPageURL = c.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			return allRepos, nil
		}
	}
}

func (c GitForge) relFeedFromRepo(
	ctx context.Context,
	repo Repo,
) ReleaseFeed {
	feedURL := fmt.Sprintf("%s/releases.atom", repo.URL)
	feed := ReleaseFeed{Name: repo.Name, URL: common.FeedURL(feedURL)}
	data, _, err := common.DoAPIRequest(
		ctx,
		http.MethodGet,
		feedURL,
		nil,
		c.headers,
		c.client,
	)
	if err != nil {
		return feed
	}

	relFeed := &AtomFeed{}
	if err = xml.Unmarshal(data, relFeed); err != nil {
		return feed
	}

	if len(relFeed.Entries) >= 1 {
		feed.Valid = true
		return feed
	}

	return feed
}

func (c GitForge) parseNextPageURL(respHeaders http.Header) string {
	linkHeader := respHeaders.Get("Link")
	if linkHeader == "" {
		return ""
	}

	links := strings.SplitSeq(linkHeader, ",")
	for link := range links {
		matches := nextPagePattern.FindStringSubmatch(link)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}

func buildHeaders(forgeType ForgeType, token string) http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("User-Agent", "github.com/atomicmeganerd/starfeed")
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if forgeType == GitHubForgeType {
		headers.Set("X-GitHub-Api-Version", "2022-11-28")
	}
	return headers
}

func buildStarredRepoUrl(forgeType ForgeType, fqdn string) string {
	if forgeType == GitHubForgeType {
		return fmt.Sprintf("https://api.%s/user/starred?per_page=100", fqdn)
	}
	return fmt.Sprintf("https://%s/api/v1/user/starred?limit=100", fqdn)
}
