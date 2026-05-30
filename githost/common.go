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
	client           *http.Client
}

func NewGitHost(
	hostCfg config.GitHostConfig, client *http.Client,
) (GitHost, error) {
	gitHost := GitHost{
		hostType: hostCfg.Type,
		Name:     hostCfg.Name,
		baseURL:  hostCfg.BaseURL,
		token:    hostCfg.Token,
		Enabled:  hostCfg.Enabled,
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
// It returns a map of releaseFeedUrl -> Repo
func (g GitHost) GetStarredRepos(
	ctx context.Context,
) (map[string]StarredRepo, error) {
	allFeeds := make(map[string]StarredRepo)
	slog.Debug("Querying git host for starred repos", "host", g.Name, "url", g.getReposURL)

	nextPageUrl := g.getReposURL
	for {
		resp, err := DoApiRequest(
			ctx,
			nextPageUrl,
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

		slog.Info(
			"Successfully loaded starred repos from Git host",
			"gitHost", g.Name,
			"numberStarredRepos", len(repos),
		)

		for _, repo := range repos {
			allFeeds[repo.FeedURL()] = repo
		}

		// If there is no next page we are done...
		if resp.NextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", resp.NextPage)
		nextPageUrl = resp.NextPage
	}
}

// This function returns true if the given repoUrl is a release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (g GitHost) IsReleaseFeedForCurrentHost(feedUrl string) bool {
	return g.isReleasePattern.MatchString(feedUrl)
}

// This function will parse different kinds of repos based on type
func (g GitHost) parseRepos(data []byte) ([]StarredRepo, error) {
	if g.hostType == "github" {
		repoSlice := make([]StarredRepo, 0)
		if err := json.Unmarshal(data, &repoSlice); err != nil {
			return nil, err
		}
		return repoSlice, nil

	}

	if g.hostType == "forgejo" {
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
	}

	return nil, fmt.Errorf("unknown hostType %s", g.hostType)
}
