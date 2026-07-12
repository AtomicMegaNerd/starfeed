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

	"github.com/atomicmeganerd/starfeed/common"
)

// This regex will match if there is a next page in the response headers
var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

type GitForge struct {
	name             string
	fetchRepoURL     string
	feeds            map[string]string
	isReleasePattern *regexp.Regexp
	headers          http.Header
	logger           *slog.Logger
	client           *http.Client
}

func NewGitForge(
	cfg GitForgeConfig,
	logger *slog.Logger,
	client *http.Client,
) *GitForge {
	return &GitForge{
		name:         cfg.Name,
		fetchRepoURL: buildStarredRepoUrl(cfg),
		feeds:        make(map[string]string, 0),
		isReleasePattern: regexp.MustCompile(
			fmt.Sprintf(
				`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(cfg.URL),
			),
		),
		headers: buildHeaders(cfg),
		logger: logger.With(
			slog.Group("gitforge",
				"name", cfg.Name,
				"type", cfg.Type,
			),
		),
		client: client,
	}
}

func (g *GitForge) LoadFeeds(
	ctx context.Context,
) error {
	// Clear the feeds map before reloading...
	g.feeds = make(map[string]string, 0)
	repos, err := g.fetchStarredRepos(ctx)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if !g.repoHasReleaseFeed(ctx, repo) {
			continue
		}
		g.feeds[repo.FeedURL] = repo.Name
	}
	g.logger.Info(
		"Successfully loaded starred repos with release feeds from Git host",
	)
	return nil
}

func (g *GitForge) fetchStarredRepos(
	ctx context.Context,
) ([]StarredRepo, error) {
	g.logger.Debug("Fetching starred repos", "url", g.fetchRepoURL)
	allRepos := make([]StarredRepo, 0)
	nextPageURL := g.fetchRepoURL
	for {
		data, respHeaders, err := common.DoAPIRequest(
			ctx,
			http.MethodGet,
			nextPageURL,
			nil,
			g.headers,
			g.client,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"error %w getting raw data from gitforge: %s url: %s",
				err, g.name, nextPageURL,
			)
		}

		repos := make([]StarredRepo, 0)
		if err := json.Unmarshal(data, &repos); err != nil {
			return nil, fmt.Errorf(
				"error %w parsing JSON response from gitforge %s",
				err, g.name,
			)
		}

		for i := range repos {
			repos[i].FeedURL = fmt.Sprintf(
				"%s/releases.atom", repos[i].RepoURL,
			)
		}
		allRepos = append(allRepos, repos...)

		nextPageURL = g.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			return allRepos, nil
		}
		g.logger.Debug("Found next page", "url", nextPageURL)
	}
}

func (g *GitForge) Feeds() map[string]string {
	return g.feeds
}

func (g *GitForge) Name() string {
	return g.name
}

func (g *GitForge) IsRepoFeedStale(feedUrl string) bool {
	// First of all, if the repo exists it canot be stale
	if _, exists := g.feeds[feedUrl]; exists {
		return false
	}

	// If the repo does not exist but matches the regex for this gitforge it is stale
	return g.isReleasePattern.MatchString(feedUrl)
}

func (g *GitForge) repoHasReleaseFeed(
	ctx context.Context,
	repo StarredRepo,
) bool {
	data, _, err := common.DoAPIRequest(ctx, http.MethodGet, repo.FeedURL, nil, g.headers, g.client)
	if err != nil {
		return false
	}
	relFeed := &AtomFeed{}
	if err = xml.Unmarshal(data, relFeed); err != nil {
		return false
	}
	if len(relFeed.Entries) >= 1 {
		return true
	}
	return false
}

func (g *GitForge) parseNextPageURL(respHeaders http.Header) string {
	linkHeader := respHeaders.Get("Link")
	if linkHeader == "" {
		return ""
	}

	g.logger.Debug("linkHeader found", "linkHeader", linkHeader)
	links := strings.SplitSeq(linkHeader, ",")
	for link := range links {
		matches := nextPagePattern.FindStringSubmatch(link)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}

func buildHeaders(cfg GitForgeConfig) http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("User-Agent", "github.com/atomicmeganerd/starfeed")
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.Token))
	if cfg.Type == GitHubForgeType {
		headers.Set("X-GitHub-Api-Version", "2022-11-28")
	}
	return headers
}

func buildStarredRepoUrl(cfg GitForgeConfig) string {
	if cfg.Type == GitHubForgeType {
		return fmt.Sprintf("%s/user/starred?per_page=100", cfg.URL)
	}
	return fmt.Sprintf("%s/user/starred?limit=100", cfg.URL)
}
