package githost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/config"
)

var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

// This object represents a supported git host where we have 'starred' repos.
type GitHost struct {
	Name     string
	Enabled  bool
	hostType string
	baseURL  string

	// These are computed
	getReposURL      string
	headers          http.Header
	nextPagePattern  *regexp.Regexp
	isReleasePattern *regexp.Regexp
	logger           *slog.Logger
	client           *http.Client
}

func NewGitHost(
	hostCfg config.GitHostConfig, logger *slog.Logger, client *http.Client,
) (GitHost, error) {
	gitHost := GitHost{
		Name:     hostCfg.Name,
		Enabled:  hostCfg.Enabled,
		hostType: hostCfg.Type,
		baseURL:  hostCfg.BaseURL,
		logger:   logger.With("githost", hostCfg.Name),
		client:   client,
	}

	// Add the common headers
	gitHost.headers = http.Header{}
	gitHost.headers.Set("Content-Type", "application/json")
	gitHost.headers.Set("Accept", "application/json")
	gitHost.headers.Set("User-Agent", "github.com/atomicmeganerd/starfeed")
	gitHost.headers.Set("Authorization", fmt.Sprintf("Bearer %s", hostCfg.Token))
	gitHost.getReposURL = fmt.Sprintf("%s/user/starred?limit=100", hostCfg.ApiURL)

	gitHost.isReleasePattern = regexp.MustCompile(
		fmt.Sprintf(
			`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
			regexp.QuoteMeta(gitHost.baseURL),
		),
	)

	// Some of the fields on this object depend on what type of git host this is...
	switch gitHost.hostType {
	case config.GitHubHostType:
		gitHost.headers.Set("X-GitHub-Api-Version", "2022-11-28")
		return gitHost, nil
	case config.ForgejoHostType:
		return gitHost, nil
	}
	return GitHost{}, errors.New("unable to build GitHostConfig")
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of releaseFeedURL -> Repo
func (g GitHost) GetStarredRepos(
	ctx context.Context,
) (map[string]StarredRepo, error) {
	allFeeds := make(map[string]StarredRepo)
	g.logger.Debug("Querying git host for starred repos", "url", g.getReposURL)

	nextPageURL := g.getReposURL
	for {
		data, respHeaders, err := common.DoAPIRequest(
			ctx,
			"GET",
			nextPageURL,
			nil,
			g.headers,
			g.client,
		)
		if err != nil {
			return nil, err
		}

		resp, err := g.processGitHostResponse(data, respHeaders)
		if err != nil {
			return nil, err
		}

		// This will get our repos based on what type they are (github, forgejo, etc.)
		repos, err := g.parseRepos(resp.Data)
		if err != nil {
			return nil, err
		}

		g.logger.Info(
			"Successfully loaded starred repos from Git host", "numberStarredRepos", len(repos),
		)

		for _, repo := range repos {
			allFeeds[repo.FeedURL()] = repo
		}

		// If there is a next page keep going
		nextPageURL = g.getNextPageURL(respHeaders)
		if nextPageURL == "" {
			return allFeeds, nil
		}
		g.logger.Debug("Found next page", "url", resp.NextPage)
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

// This function will parse different kinds of repos based on type
func (g GitHost) parseRepos(data []byte) ([]StarredRepo, error) {
	switch g.hostType {
	case config.GitHubHostType:
		repoSlice := make([]StarredRepo, 0)
		if err := json.Unmarshal(data, &repoSlice); err != nil {
			return nil, err
		}
		return repoSlice, nil
	case config.ForgejoHostType:
		forgejoSlice := make([]forgejoRepo, 0)
		if err := json.Unmarshal(data, &forgejoSlice); err != nil {
			return nil, err
		}

		repoSlice := make([]StarredRepo, 0)
		for _, repo := range forgejoSlice {
			if repo.HasReleases {
				repoSlice = append(repoSlice, repo.StarredRepo)
			}
		}
		return repoSlice, nil
	default:
		return nil, fmt.Errorf("unknown hostType %s", g.hostType)
	}
}

func (g GitHost) processGitHostResponse(
	data []byte,
	respHeaders http.Header,
) (*GitHostResponse, error) {
	links := strings.SplitSeq(respHeaders.Get("Link"), ",")
	for link := range links {
		matches := nextPagePattern.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GitHostResponse{Data: data, NextPage: matches[1]}, nil
		}
	}

	return &GitHostResponse{Data: data}, nil
}

func (g GitHost) getNextPageURL(respHeaders http.Header) string {
	nextPage := ""
	links := strings.SplitSeq(respHeaders.Get("Link"), ",")
	for link := range links {
		matches := g.nextPagePattern.FindStringSubmatch(link)
		if len(matches) == 2 {
			nextPage = matches[1]
		}
	}
	return nextPage
}
