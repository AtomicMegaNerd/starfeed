package githost

import (
	"context"
	"encoding/json"
	"encoding/xml"
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
type GitHost struct {
	Name    string
	Enabled bool

	hostType         string
	getReposURL      string
	headers          http.Header
	isReleasePattern *regexp.Regexp
	logger           *slog.Logger
	client           *http.Client
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
		Name:        hostCfg.Name,
		Enabled:     hostCfg.Enabled,
		hostType:    hostCfg.Type,
		getReposURL: fmt.Sprintf("%s/user/starred?limit=50", hostCfg.ApiURL),
		headers:     headers,
		// This pattern does have to match for each instance
		isReleasePattern: regexp.MustCompile(
			fmt.Sprintf(
				`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(hostCfg.BaseURL),
			),
		),
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
	g.logger.Debug("Querying git host for starred repos", "url", g.getReposURL)
	nextPageURL := g.getReposURL
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

		repos, err := g.parseRepos(data)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			err := g.addReleaseFeedToRepo(ctx, &repo)
			if err != nil {
				return nil, fmt.Errorf(
					"error %w adding release feeds to repo %s from githost %s",
					err, repo.Name, g.Name,
				)
			}
		}

		// Delete all feeds that do not have a feed URL
		repos = slices.DeleteFunc(repos, func(repo StarredRepo) bool {
			return repo.FeedURL == ""
		})

		// A map makes everything easy to search based on feed
		repoFeedMap := make(map[string]StarredRepo, len(repos))
		for _, repo := range repos {
			repoFeedMap[repo.FeedURL] = repo
		}

		// If there is a next page keep going
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

// We never want to unsubscribe from feeds that are not release feeds for the current Git host.
func (g GitHost) filterOutNonRepoReleaseFeeds(
	rssFeedSet map[string]StarredRepo,
) map[string]StarredRepo {
	filteredSet := make(map[string]StarredRepo)
	for feedURL, repo := range rssFeedSet {
		// This will only include a feed for potential removal if it is a release feed
		// for the current GitHost that we are working with. This is important otherwise
		// we could remove feeds from other Git hosts which we do not want...
		if g.isReleasePattern.MatchString(feedURL) {
			filteredSet[feedURL] = repo
		} else {
			g.logger.Debug(
				"Ignoring feeds that aren't release feeds so we don't unsubscribe",
				"feed", feedURL,
			)
		}
	}
	return filteredSet
}

func (g GitHost) parseNextPageURL(respHeaders http.Header) string {
	links := strings.SplitSeq(respHeaders.Get("Link"), ",")
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

func (g GitHost) addReleaseFeedToRepo(
	ctx context.Context,
	repo *StarredRepo,
) error {
	feedURL := fmt.Sprintf("%s/releases.atom", repo.RepoURL)
	if feedURL == "" {
		return nil
	}

	data, _, err := common.DoAPIRequest(ctx, http.MethodGet, feedURL, nil, g.headers, g.client)
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
	repo.FeedURL = feedURL
	return nil
}

func buildCommonHeaders(token string) http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("User-Agent", "github.com/atomicmeganerd/starfeed")
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return headers
}
