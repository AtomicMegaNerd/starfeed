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

// This object represents a supported git host where we have 'starred' repos.
type GitForge struct {
	name         string
	fetchRepoURL string
	feedRepoMap  common.FeedRepoMap
	headers      http.Header
	logger       *slog.Logger
	client       *http.Client
}

func NewGitForge(
	cfg GitForgeConfig,
	logger *slog.Logger,
	client *http.Client,
) *GitForge {
	return &GitForge{
		name:         cfg.Name,
		headers:      buildHeaders(cfg),
		fetchRepoURL: buildStarredRepoUrl(cfg),
		feedRepoMap:  make(common.FeedRepoMap, 0),
		logger: logger.With(
			slog.Group("gitforge",
				"name", cfg.Name,
				"type", cfg.Type,
			),
		),
		client: client,
	}
}

func (g *GitForge) LoadRepoMap(
	ctx context.Context,
) error {
	g.logger.Debug("Loading feeds for starred repos", "url", g.fetchRepoURL)
	nextPageURL := g.fetchRepoURL
	for {
		// Get the raw data
		data, respHeaders, err := common.DoAPIRequest(
			ctx,
			http.MethodGet,
			nextPageURL,
			nil,
			g.headers,
			g.client,
		)
		if err != nil {
			return fmt.Errorf(
				"error %w getting raw data from gitforge: %s url: %s", err, g.name, nextPageURL,
			)
		}

		// Parse Repos
		repos := make([]StarredRepo, 0)
		if err := json.Unmarshal(data, &repos); err != nil {
			return fmt.Errorf(
				"error %w parsing JSON response from gitforge %s", err, g.name,
			)
		}

		for _, repo := range repos {
			repo.FeedURL = fmt.Sprintf("%s/releases.atom", repo.RepoURL)
			if !g.repoHasRelaseFeed(ctx, repo) {
				continue
			}
			g.feedRepoMap[repo.FeedURL] = repo.Name
		}

		nextPageURL = g.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			g.logger.Info("Successfully loaded starred repos with release feeds from Git host")
			return nil
		}

		g.logger.Debug("Found next page", "url", nextPageURL)
	}
}

func (g *GitForge) FeedRepoMap() common.FeedRepoMap {
	return g.feedRepoMap
}

func (g *GitForge) Name() string {
	return g.name
}

func (g *GitForge) repoHasRelaseFeed(
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
		return fmt.Sprintf("%s/user/starred?per_page=100", cfg.ApiURL)
	}
	return fmt.Sprintf("%s/user/starred?limit=100", cfg.ApiURL)
}
