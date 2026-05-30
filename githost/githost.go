package githost

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/atomicmeganerd/starfeed/config"
)

// This object represents a supported git host where we have 'starred' repos.
type GitHost struct {
	Name    string
	Enabled bool

	hostType string
	baseURL  string
	token    string

	// These are computed
	getReposURL      string
	headers          map[string]string
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
		token:    hostCfg.Token,
		logger:   logger.With("githost", hostCfg.Name),
		client:   client,
	}

	// Some of the fields on this object depend on what type of git host this is...
	switch gitHost.hostType {
	case "github":
		gitHost.headers = map[string]string{
			"Authorization":        fmt.Sprintf("Bearer %s", gitHost.token),
			"X-GitHub-Api-Version": "2022-11-28",
			"User-Agent":           "github.com/atomicmeganerd/starfeed",
			"Content-Type":         "application/json",
			"Accept":               "application/json",
		}
		gitHost.getReposURL = fmt.Sprintf("%s/user/starred?per_page=100", hostCfg.ApiURL)

		gitHost.nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)
		gitHost.isReleasePattern = regexp.MustCompile(
			fmt.Sprintf(
				`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(gitHost.baseURL),
			),
		)

		return gitHost, nil

	case "forgejo":
		gitHost.headers = map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", gitHost.token),
			"User-Agent":    "github.com/atomicmeganerd/starfeed",
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		}

		gitHost.nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)
		gitHost.isReleasePattern = regexp.MustCompile(
			fmt.Sprintf(
				`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
				regexp.QuoteMeta(gitHost.baseURL),
			),
		)

		gitHost.getReposURL = fmt.Sprintf("%s/user/starred?limit=50", hostCfg.ApiURL)
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
		resp, err := DoAPIRequest(
			ctx,
			nextPageURL,
			g.headers,
			g.nextPagePattern,
			g.client,
		)
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

		// If there is no next page we are done...
		if resp.NextPage == "" {
			return allFeeds, nil
		}

		g.logger.Debug("Found next page", "url", resp.NextPage)
		nextPageURL = resp.NextPage
	}
}

// We never want to unsubscribe from feeds that are not release feeds for the current Git host.
func (g GitHost) FilterOutNonRepoReleaseFeeds(
	rssFeedSet map[string]struct{},
) map[string]struct{} {
	// NOTE: In Go map[T]struct{} is the idiomatic way to make a set as struct{} is 0-bytes
	filteredSet := make(map[string]struct{})
	for feedURL := range rssFeedSet {
		// This will only include a feed for potential removal if it is a release feed
		// for the current GitHost that we are working with. This is important otherwise
		// we could remove feeds from other Git hosts which we do not want...
		if g.isReleasePattern.MatchString(feedURL) {
			filteredSet[feedURL] = struct{}{}
		} else {
			g.logger.Debug(
				"Ignoring feeds that aren't release feeds from a git host so we don't unsubscribe",
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
