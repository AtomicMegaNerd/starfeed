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
	Enabled() bool
	GetStarredRepos(context.Context) (map[string]Repo, error)
	IsReleaseFeedForCurrentHost(string) bool
}

// This object represents a supported git host where we have 'starred' repos.
type gitHost struct {
	hostType string
	name     string
	baseURL  string
	token    string
	enabled  bool

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
		enabled:  hostCfg.Enabled,
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

	return nil, errors.New("unable to build GitHostConfig")
}

func (g *gitHost) Name() string {
	return g.name
}

func (g *gitHost) Enabled() bool {
	return g.enabled
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

		// This will get our repos based on what type they are (github, forgejo, etc.)
		repos, err := parseRepos(resp.Data, g.hostType)
		if err != nil {
			return nil, err
		}

		slog.Info(
			"Sucessfully loaded starred repos from Git host",
			"gitHost", g.Name(),
			"numberStarredRepos", len(repos),
		)

		for _, repo := range repos {
			feedUrl := repo.FeedURL()
			if feedUrl == "" {
				slog.Debug(
					"Skipping repo without releases",
					"gitHost", g.Name(),
					"repo", repo.Name(),
				)
				continue
			}

			slog.Debug(
				"Parsed starred repo from JSON",
				"gitHost", g.Name,
				"repo", repo.Name(),
				"kind", g.hostType,
				"feedUrl", feedUrl,
			)

			allFeeds[feedUrl] = repo
		}

		// If there is no next page we are done...
		if resp.NextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", resp.NextPage)
		nextPageUrl = resp.NextPage
	}
}

// This generic method will parse any supported repo type
func parseTypedRepos[T BaseRepo | ForgejoRepo](data []byte) ([]Repo, error) {
	// This will be the concrete type so that JSON can unmarshal
	var genericRepoSlice []T
	if err := json.Unmarshal(data, &genericRepoSlice); err != nil {
		return nil, err
	}

	// Now to make this work with the rest of the app we have to copy the concrete types into a
	// slice of the interface type
	repoInterfaceSlice := make([]Repo, len(genericRepoSlice))
	for i := range genericRepoSlice {
		repo := genericRepoSlice[i]
		// This forces a runtime check to ensure that r always satisfies the Repo interface
		// this will panic if we are not correct, but both concrete types of T do so we are fine.
		repoInterfaceSlice[i] = any(&repo).(Repo)
	}

	return repoInterfaceSlice, nil
}

func parseRepos(data []byte, hostType string) ([]Repo, error) {
	switch hostType {
	case "github":
		return parseTypedRepos[BaseRepo](data)
	case "forgejo":
		return parseTypedRepos[ForgejoRepo](data)
	}
	return nil, fmt.Errorf("unknown host type: %s", hostType)
}

// This function returns true if the given repoUrl is a release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (g *gitHost) IsReleaseFeedForCurrentHost(feedUrl string) bool {
	return g.isReleasePattern.MatchString(feedUrl)
}
