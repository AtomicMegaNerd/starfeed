package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

// This object handles buidling an Atom Feed of all starred repos for the authenticated
// user.
type GitHubStarredFeedBuilder struct {
	token  string // WARNING: Do not log this value as it is a secret
	ctx    context.Context
	client *http.Client
	re     *regexp.Regexp
}

func NewGitHubStarredFeedBuilder(
	token string,
	ctx context.Context,
	client *http.Client,
) *GitHubStarredFeedBuilder {

	return &GitHubStarredFeedBuilder{token: token, ctx: ctx, client: client}
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of relaseFeedUrl -> GitHubRepo
func (gh *GitHubStarredFeedBuilder) GetStarredRepos() (map[string]GitHubRepo, error) {
	allFeeds := make(map[string]GitHubRepo)
	getUrl := "http://api.github.com/user/starred?per_page=100"
	slog.Debug("Querying Github for starred repos", "url", getUrl)

	pattern := `<([^>]+)>; rel="next"`
	var err error
	gh.re, err = regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	for {
		ghResponse, err := gh.doApiRequest(getUrl)
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		if err = json.Unmarshal(ghResponse.data, &repos); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			repo.BuildReleasesFeedURL()
			allFeeds[repo.FeedUrl] = repo
		}

		// If there is no next page we are done...
		if ghResponse.nextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", ghResponse.nextPage)
		getUrl = ghResponse.nextPage
	}
}

// This method handles making API requests to GitHub using its REST API.
func (gh *GitHubStarredFeedBuilder) doApiRequest(url string) (*GithubResponse, error) {
	headers := map[string]string{
		"Authorization":        fmt.Sprintf("Bearer %s", gh.token),
		"X-Github-Api-Version": "2022-11-28",
		"User-Agent":           "github.com/atomicmeganerd/starfeed",
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}

	httpRequest, err := http.NewRequestWithContext(gh.ctx, "GET", url, nil)
	if err != nil {
		slog.Error("Unable to build request to Github", "error", err.Error())
		return nil, err
	}

	for k, v := range headers {
		httpRequest.Header.Set(k, v)
	}

	res, err := gh.client.Do(httpRequest)
	if err != nil {
		slog.Error("Unable to make request to Github", "error", err.Error())
		return nil, err

	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned an http error code %d", res.StatusCode)
	}

	ghResponse, err := gh.processGithubResponse(res)
	if err != nil {
		slog.Error("Unable to parse response from Github", "error", err)
		return nil, err
	}

	return ghResponse, nil
}

// This function processes the response from GitHub. It will both read
// the data from the response and check for a next page if there is
// one
func (gh *GitHubStarredFeedBuilder) processGithubResponse(r *http.Response) (*GithubResponse, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	linkRaw := r.Header.Get("link")
	links := strings.Split(linkRaw, ",")
	for _, link := range links {
		matches := gh.re.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GithubResponse{data: data, nextPage: matches[1]}, nil
		}
	}

	return &GithubResponse{data: data}, nil
}
