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

type GitHost interface {
	Name() string
	GetStarredRepos(context.Context) (map[string]Repo, error)
	IsReleaseFeed(string) bool
}

// This is the config for each Git Host that we read from the environment. We read it from an
// environment variable called `STARFEED_GIT_HOST_ix` where ix is a number from 0..n.
// The format of the CSV is as follows:
// type,name,url,token
type gitHost struct {
	hostType string
	name     string
	baseURL  string
	token    string

	// These are computed
	GetReposURL      string
	Headers          map[string]string
	NextPagePattern  *regexp.Regexp
	IsReleasePattern *regexp.Regexp
	client           *http.Client
}

func NewGitHost(
	hostCfg config.GitHostConfig, client *http.Client,
) (GitHost, error) {
	gitHost := &gitHost{
		hostType: hostCfg.Type,
		name:     hostCfg.Name,
		baseURL:  hostCfg.BaseURL,
		token:    hostCfg.Token,
		client:   client,
	}

	// This regex is used to find the next page link in the GitHub API response
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	// This regex is used to determine if an RSS feed is a Forgejo release feed
	isRelRepoRegex := regexp.MustCompile(
		fmt.Sprintf(`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`, regexp.QuoteMeta(gitHost.baseURL)),
	)

	gitHost.NextPagePattern = nextPageLinkRegex
	gitHost.IsReleasePattern = isRelRepoRegex

	switch gitHost.hostType {
	case "github":
		gitHost.Headers = map[string]string{
			"Authorization":        fmt.Sprintf("Bearer %s", gitHost.token),
			"X-GitHub-Api-Version": "2022-11-28",
			"User-Agent":           "github.com/atomicmeganerd/starfeed",
			"Content-Type":         "application/json",
			"Accept":               "application/json",
		}
		gitHost.GetReposURL = ""
		return gitHost, nil

	case "forgejo":
		gitHost.Headers = map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", gitHost.token),
			"User-Agent":    "TBD", // TODO: Fix this
			"Content-Type":  "application/json",
			"Accept":        "application/json",
		}
		gitHost.GetReposURL = ""
		return gitHost, nil
	}

	return nil, errors.New("unable to build GitHostConfig")
}

func (g *gitHost) Name() string {
	return g.name
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of releaseFeedUrl -> GitHubRepo
func (g *gitHost) GetStarredRepos(
	ctx context.Context,
) (map[string]Repo, error) {
	allFeeds := make(map[string]Repo)
	getUrl := "https://api.github.com/user/starred?per_page=100"
	slog.Debug("Querying GitHub for starred repos", "url", getUrl)

	for {
		resp, err := DoApiRequest(
			ctx,
			getUrl,
			g.Headers,
			g.NextPagePattern,
			g.client,
		)
		if err != nil {
			return nil, err
		}

		var repos []BaseRepo
		if err = json.Unmarshal(resp.Data, &repos); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allFeeds[repo.FeedURL()] = &repo
		}

		// If there is no next page we are done...
		if resp.NextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", resp.NextPage)
		getUrl = resp.NextPage
	}
}

// This function returns true if a repoUrl is a GitHub release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (g *gitHost) IsReleaseFeed(feedUrl string) bool {
	return g.IsReleasePattern.MatchString(feedUrl)
}
