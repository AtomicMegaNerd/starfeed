package gitforge

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/atomicmeganerd/starfeed/common"
	"golang.org/x/sync/errgroup"
)

// This regex will match if there is a next page in the response headers
var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

// GitForgeClient struct represents a GitForge. We can load RSS feeds for all starred repos that
// belong to this Git Forge.
type GitForgeClient struct {
	fetchRepoURL string
	headers      http.Header
	logger       *slog.Logger
	client       *http.Client
}

func NewGitForgeClient(
	forgeType, fqdn, token string,
	logger *slog.Logger,
	client *http.Client,
) GitForgeClient {
	return GitForgeClient{
		fetchRepoURL: buildStarredRepoUrl(forgeType, fqdn),
		headers:      buildHeaders(forgeType, token),
		logger:       logger,
		client:       client,
	}
}

func (c GitForgeClient) LoadFeeds(
	ctx context.Context,
) (FeedResultMap, error) {
	starredRepos, err := c.fetchStarredRepos(ctx)
	if err != nil {
		return nil, err
	}

	starredFeeds := make(FeedResultMap)

	// Check each repo to make sure it has valid entries in its ATOM feed for releases
	// This can be done in parallel to make it much faster.
	mu := sync.Mutex{}
	// We only use a errgroup here to get SetLimit. None of our goroutines can throw an
	// error. I just like this better than using the weighted semaphore.
	eg := &errgroup.Group{}
	eg.SetLimit(5)
	for _, repo := range starredRepos {
		logger := c.logger.With("repoName", repo.Name, "feeedURL", repo.FeedURL)
		eg.Go(func() error {
			result := c.repoHasReleaseFeed(ctx, repo)
			if result.IsOK() {
				logger.Info("Repo has valid release feed")
			}
			mu.Lock()
			starredFeeds[repo.FeedURL] = result
			mu.Unlock()
			return nil
		})
	}

	_ = eg.Wait()

	c.logger.Info("Successfully added all feeds to feeds map", "numFeeds", len(starredFeeds))
	return starredFeeds, nil
}

func (c GitForgeClient) fetchStarredRepos(
	ctx context.Context,
) ([]GitRepo, error) {
	allRepos := make([]GitRepo, 0)
	nextPageURL := c.fetchRepoURL
	for {
		c.logger.Debug("Fetching starred repos", "url", nextPageURL)
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

		repos := make([]GitRepo, 0)
		if err := json.Unmarshal(data, &repos); err != nil {
			return nil, fmt.Errorf(
				"error %w parsing JSON response from gitforge", err,
			)
		}

		for ix := range repos {
			repos[ix].FeedURL = common.FeedURL(
				fmt.Sprintf(
					"%s/releases.atom", repos[ix].RepoURL,
				),
			)
		}
		allRepos = append(allRepos, repos...)

		nextPageURL = c.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			c.logger.Info("Finished loading starred repos", "numRepos", len(allRepos))
			return allRepos, nil
		}
	}
}

func (c GitForgeClient) repoHasReleaseFeed(
	ctx context.Context,
	repo GitRepo,
) GitRepoResult {

	result := GitRepoResult{RepoName: repo.Name}

	logger := c.logger.With("repo", repo.Name, "feed", repo.FeedURL)
	logger.Debug("Checking if repo has release feed")
	data, _, err := common.DoAPIRequest(
		ctx,
		http.MethodGet,
		repo.FeedURL.String(),
		nil,
		c.headers,
		c.client,
	)
	if err != nil {
		result.Err = err
		return result
	}
	relFeed := &AtomFeed{}
	if err = xml.Unmarshal(data, relFeed); err != nil {
		result.Err = err
		return result
	}
	if len(relFeed.Entries) >= 1 {
		logger.Debug("Repo feed is valid")
		result.RelFeedHasEntries = true
		return result
	}

	return result
}

func (c GitForgeClient) parseNextPageURL(respHeaders http.Header) string {
	linkHeader := respHeaders.Get("Link")
	if linkHeader == "" {
		return ""
	}

	c.logger.Debug("linkHeader found", "linkHeader", linkHeader)
	links := strings.SplitSeq(linkHeader, ",")
	for link := range links {
		matches := nextPagePattern.FindStringSubmatch(link)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}

func buildHeaders(forgeType, token string) http.Header {
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

func buildStarredRepoUrl(forgeType, fqdn string) string {
	if forgeType == GitHubForgeType {
		return fmt.Sprintf("https://api.%s/user/starred?per_page=100", fqdn)
	}
	return fmt.Sprintf("https://%s/api/v1/user/starred?limit=100", fqdn)
}
