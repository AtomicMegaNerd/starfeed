package gitforge

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/atomicmeganerd/starfeed/common"
	"golang.org/x/sync/errgroup"
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
	mtx              sync.RWMutex
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
				`^https://%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(cfg.Fqdn),
			),
		),
		headers: buildHeaders(cfg),
		logger: logger.With(
			slog.Group("gitforge",
				"name", cfg.Name,
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

	// Get all repos
	repos, err := g.fetchStarredRepos(ctx)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(5)

	// Check each repo to make sure it has valid entries in its ATOM feed for releases
	// This can be done in parallel
	for _, repo := range repos {
		eg.Go(func() error {
			logger := g.logger.With(
				"feed", repo.FeedURL,
				"repo", repo.Name,
			)
			if !g.repoHasReleaseFeed(ctx, repo) {
				logger.Warn("Repo does not have valid release feed")
				return nil
			}

			g.mtx.Lock()
			g.feeds[repo.FeedURL] = repo.Name
			g.mtx.Unlock()
			g.logger.Info("Added feed for repo to feeds map")
			return nil
		})
	}
	// We don't get an error
	_ = eg.Wait()

	g.logger.Info(
		"Successfully added all feeds to feeds map",
		"numFeeds", len(g.feeds),
	)
	return nil
}

func (g *GitForge) fetchStarredRepos(
	ctx context.Context,
) ([]GitRepo, error) {
	allRepos := make([]GitRepo, 0)
	nextPageURL := g.fetchRepoURL
	for {
		g.logger.Debug("Fetching starred repos", "url", nextPageURL)
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

		repos := make([]GitRepo, 0)
		if err := json.Unmarshal(data, &repos); err != nil {
			return nil, fmt.Errorf(
				"error %w parsing JSON response from gitforge %s",
				err, g.name,
			)
		}

		for ix := range repos {
			repos[ix].FeedURL = fmt.Sprintf(
				"%s/releases.atom", repos[ix].RepoURL,
			)
		}
		allRepos = append(allRepos, repos...)

		nextPageURL = g.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			g.logger.Info("Finished loading starred repos", "numRepos", len(allRepos))
			return allRepos, nil
		}
	}
}

func (g *GitForge) Feeds() map[string]string {
	g.mtx.RLock()
	defer g.mtx.RUnlock()
	feedsCopy := maps.Clone(g.feeds)
	return feedsCopy
}

func (g *GitForge) Name() string {
	return g.name
}

func (g *GitForge) IsRepoFeedStale(feedUrl string) bool {
	// First of all, if the repo exists it canot be stale
	g.mtx.RLock()
	if _, exists := g.feeds[feedUrl]; exists {
		return false
	}
	g.mtx.RUnlock()

	// If the repo does not exist but matches the regex for this gitforge it is stale
	return g.isReleasePattern.MatchString(feedUrl)
}

func (g *GitForge) repoHasReleaseFeed(
	ctx context.Context,
	repo GitRepo,
) bool {
	logger := g.logger.With("repo", repo.Name, "feed", repo.FeedURL)
	logger.Debug("Checking if repo has release feed")
	data, _, err := common.DoAPIRequest(ctx, http.MethodGet, repo.FeedURL, nil, g.headers, g.client)
	if err != nil {
		return false
	}
	relFeed := &AtomFeed{}
	if err = xml.Unmarshal(data, relFeed); err != nil {
		return false
	}
	if len(relFeed.Entries) >= 1 {
		logger.Debug("Repo feed is valid")
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
		return fmt.Sprintf("https://api.%s/user/starred?per_page=100", cfg.Fqdn)
	}
	return fmt.Sprintf("https://%s/api/v1/user/starred?limit=100", cfg.Fqdn)
}
