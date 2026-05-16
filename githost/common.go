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
	getReposURL      string
	headers          map[string]string
	nextPagePattern  *regexp.Regexp
	isReleasePattern *regexp.Regexp
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
		gitHost.getReposURL = "https://api.github.com/user/starred?per_page=100"

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
			"User-Agent":    "TBD", // TODO: Fix this
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

		gitHost.getReposURL = ""
		return gitHost, nil
	}

	return nil, errors.New("unable to build GitHostConfig")
}

func (g *gitHost) Name() string {
	return g.name
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of releaseFeedUrl -> Repo
func (g *gitHost) GetStarredRepos(
	ctx context.Context,
) (map[string]Repo, error) {
	allFeeds := make(map[string]Repo)
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
		nextPageUrl = resp.NextPage
	}
}

// This function returns true if the given repoUrl is a release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (g *gitHost) IsReleaseFeed(feedUrl string) bool {
	return g.isReleasePattern.MatchString(feedUrl)
}
