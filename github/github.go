package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/charmbracelet/log"
)

// This object handles buidling an Atom Feed of all starred repos for the authenticated
// user.
type GitHubStarredFeedBuilder struct {
	token  string // WARNING: Do not log this value as it is a secret
	ctx    context.Context
	client *http.Client
}

func NewGitHubStarredFeedBuilder(
	token string,
	client *http.Client,
) *GitHubStarredFeedBuilder {
	ctx := context.Background()
	return &GitHubStarredFeedBuilder{token: token, ctx: ctx, client: client}
}

// This will return all starred repos including the Atom feeds for their releases
func (gh *GitHubStarredFeedBuilder) GetStarredRepos() (map[string]string, error) {
	allFeeds := make(map[string]string)
	getUrl := "http://api.github.com/user/starred?per_page=100&fields=id,name,full_name,html_url"
	log.Debugf("Querying Github for starred repos: %s", getUrl)

	for {
		ghResponse, err := gh.doApiRequest(getUrl)
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		err = json.Unmarshal(ghResponse.data, &repos)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			repo.BuildReleasesFeedURL()
			allFeeds[repo.FeedUrl] = repo.Name
		}

		// If there is no next page we are done...
		if ghResponse.nextPage == "" {
			return allFeeds, nil
		}

		log.Debugf("Found next page: %s", ghResponse.nextPage)
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
		log.Error("Unable to build request to github", err)
		return nil, err
	}

	for k, v := range headers {
		httpRequest.Header.Set(k, v)
	}

	res, err := gh.client.Do(httpRequest)
	if err != nil {
		log.Error("Unable to make request to Github", err)
		return nil, err

	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned an http error code %d", res.StatusCode)
	}

	ghResponse, err := gh.processGithubResponse(res)
	if err != nil {
		log.Error("Unable to parse response from github", err)
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
	pattern := `<([^>]+)>; rel="next"`

	re, err := regexp.Compile(pattern)

	if err != nil {
		return nil, err
	}

	links := strings.Split(linkRaw, ",")
	for _, link := range links {
		matches := re.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GithubResponse{data: data, nextPage: matches[1]}, nil
		}
	}

	return &GithubResponse{data: data}, nil
}

func (gh *GitHubStarredFeedBuilder) CheckIfFeedHasEntries(feedUrl string) bool {
	// TODO implement this
	return false
}
