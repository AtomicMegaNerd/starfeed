package githost

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
	"github.com/atomicmeganerd/starfeed/config"
)

// This regex will match if there is a next page in the response headers
var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

// This object represents a supported git host where we have 'starred' repos.
type GitHost struct {
	Name               string
	Enabled            bool
	HostType           string
	ReleaseFeedPattern *regexp.Regexp

	starredReposFetchURL string
	headers              http.Header
	logger               *slog.Logger
	client               *http.Client
}

func NewGitHost(
	hostCfg config.GitHostConfig,
	logger *slog.Logger,
	client *http.Client,
) GitHost {
	headers := buildCommonHeaders(hostCfg.Token)
	if hostCfg.Type == config.GitHubHostType {
		headers.Set("X-GitHub-Api-Version", "2022-11-28")
	}

	return GitHost{
		Name:     hostCfg.Name,
		Enabled:  hostCfg.Enabled,
		HostType: hostCfg.Type,
		// This pattern does have to match for each instance
		ReleaseFeedPattern: regexp.MustCompile(
			fmt.Sprintf(
				`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(hostCfg.BaseURL),
			),
		),
		starredReposFetchURL: fmt.Sprintf("%s/user/starred?limit=50", hostCfg.ApiURL),
		headers:              headers,
		logger: logger.With(
			slog.Group("githost",
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
func (g GitHost) GetStarredRepos(
	ctx context.Context,
) (map[string]StarredRepo, error) {
	g.logger.Debug("Querying git host for starred repos", "url", g.starredReposFetchURL)

	// A map makes everything easy to search based on feed
	repoFeedMap := make(map[string]StarredRepo)
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
				"error %w getting raw data from githost: %s url: %s", err, g.Name, nextPageURL,
			)
		}

		// Parse Repos
		repos, err := g.parseRepos(data)
		if err != nil {
			return nil, err
		}

		g.logger.Debug("Parsed repos", "# repos", len(repos))

		for _, repo := range repos {
			repo.FeedURL = fmt.Sprintf("%s/releases.atom", repo.RepoURL)
			repoFeedMap[repo.FeedURL] = repo
		}

		nextPageURL = g.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			g.logger.Info(
				"Successfully loaded starred repos with release feeds from Git host",
				"numberStarredRepos", len(repoFeedMap),
			)
			return repoFeedMap, nil
		}

		g.logger.Debug("Found next page", "url", nextPageURL)
	}
}

func (g GitHost) CheckReleaseFeed(
	ctx context.Context,
	repo *StarredRepo,
) error {
	data, _, err := common.DoAPIRequest(ctx, http.MethodGet, repo.FeedURL, nil, g.headers, g.client)
	if err != nil {
		return err
	}

	var feed AtomFeed
	if err = xml.Unmarshal(data, &feed); err != nil {
		return err
	}

	if len(feed.Entries) < 1 {
		return nil
	}

	// Set the release feed
	return nil
}

func (g GitHost) parseNextPageURL(respHeaders http.Header) string {
	linkHeader := respHeaders.Get("Link")

	if linkHeader != "" {
		g.logger.Debug("linkHeader found", "linkHeader", linkHeader)
	} else {
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

func (g GitHost) parseRepos(data []byte) ([]StarredRepo, error) {
	repoSlice := make([]StarredRepo, 0)
	if err := json.Unmarshal(data, &repoSlice); err != nil {
		return nil, fmt.Errorf(
			"error %w parsing JSON response from githost %s", err, g.Name,
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
