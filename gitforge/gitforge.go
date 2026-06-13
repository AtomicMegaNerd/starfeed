package gitforge

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/config"
)

// This regex will match if there is a next page in the response headers
var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

// This object represents a supported git host where we have 'starred' repos.
type GitForge struct {
	Name                 string
	Enabled              bool
	HostType             string
	releaseFeedPattern   *regexp.Regexp
	starredReposFetchURL string
	headers              http.Header
	logger               *slog.Logger
	client               *http.Client
}

func NewGitForge(
	hostCfg config.GitForgeConfig,
	logger *slog.Logger,
	client *http.Client,
) GitForge {
	headers := buildCommonHeaders(hostCfg.Token)

	// The URL to fetch starred repos does differ slightly
	starredReposFetchURL := ""
	if hostCfg.Type == config.GitHubForgeType {
		headers.Set("X-GitHub-Api-Version", "2022-11-28")
		starredReposFetchURL = fmt.Sprintf("%s/user/starred?per_page=100", hostCfg.ApiURL)
	}
	if hostCfg.Type == config.ForgejoForgeType {
		starredReposFetchURL = fmt.Sprintf("%s/user/starred?limit=100", hostCfg.ApiURL)
	}

	return GitForge{
		Name:     hostCfg.Name,
		Enabled:  hostCfg.Enabled,
		HostType: hostCfg.Type,
		// This pattern has to match for each instance
		releaseFeedPattern: regexp.MustCompile(
			fmt.Sprintf(
				`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(hostCfg.BaseURL),
			),
		),
		starredReposFetchURL: starredReposFetchURL,
		headers:              headers,
		logger: logger.With(
			slog.Group("gitforge",
				"name", hostCfg.Name,
				"type", hostCfg.Type,
				"baseURL", hostCfg.BaseURL,
			),
		),
		client: client,
	}
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of releaseFeedURL -> Repo
func (g GitForge) GetStarredRepos(
	ctx context.Context,
) ([]StarredRepo, error) {
	g.logger.Debug("Querying git host for starred repos", "url", g.starredReposFetchURL)

	// A map makes everything easy to search based on feed
	starredRepos := make([]StarredRepo, 0)
	nextPageURL := g.starredReposFetchURL
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
			return nil, fmt.Errorf(
				"error %w getting raw data from gitforge: %s url: %s", err, g.Name, nextPageURL,
			)
		}

		// Parse Repos
		repos, err := g.parseRepos(data)
		if err != nil {
			return nil, err
		}

		starredRepos = slices.Concat(starredRepos, repos)
		g.logger.Debug("Parsed repos", "# repos", len(repos))

		nextPageURL = g.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			g.logger.Info(
				"Successfully loaded starred repos with release feeds from Git host",
				"numberStarredRepos", len(starredRepos),
			)
			return starredRepos, nil
		}

		g.logger.Debug("Found next page", "url", nextPageURL)
	}
}

// This method checks the atom feed for a release repo. If it finds at least one entry
// it them sets the FeedURL on the repo.
func (g GitForge) CheckReleaseFeedExistsAndHasEntries(
	ctx context.Context,
	repo *StarredRepo,
) error {
	if repo.RepoURL == "" {
		return fmt.Errorf("repoURL empty for repo %s", repo.Name)
	}

	feedURL := fmt.Sprintf("%s/releases.atom", repo.RepoURL)

	data, _, err := common.DoAPIRequest(ctx, http.MethodGet, feedURL, nil, g.headers, g.client)
	if err != nil {
		// If the release feed is simply not found don't return an error
		httpErr := common.HTTPError{}
		if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
			g.logger.Debug("repo does not have release feed, skipping without error")
			return nil
		}

		return err
	}

	feed := &AtomFeed{}
	if err = xml.Unmarshal(data, feed); err != nil {
		return err
	}

	if len(feed.Entries) >= 1 {
		// Set the release feed
		repo.FeedURL = feedURL
		g.logger.Debug("repo has releases", "repo", repo.Name, "feed", repo.FeedURL)
	}

	return nil
}

func (g GitForge) IsReleaseeFeedForThisHost(rssFeed string) bool {
	return g.releaseFeedPattern.MatchString(rssFeed)
}

func (g GitForge) parseNextPageURL(respHeaders http.Header) string {
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

func (g GitForge) parseRepos(data []byte) ([]StarredRepo, error) {
	repoSlice := make([]StarredRepo, 0)
	if err := json.Unmarshal(data, &repoSlice); err != nil {
		return nil, fmt.Errorf(
			"error %w parsing JSON response from gitforge %s", err, g.Name,
		)
	}
	return repoSlice, nil
}

func buildCommonHeaders(token string) http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("User-Agent", "github.com/atomicmeganerd/starfeed")
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return headers
}
