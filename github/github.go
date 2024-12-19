package github

import (
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
	client *http.Client
	logger *log.Logger
}

func NewGitHubStarredFeedBuilder(
	token string,
	client *http.Client,
	logger *log.Logger,
) *GitHubStarredFeedBuilder {
	return &GitHubStarredFeedBuilder{token: token, client: client, logger: logger}
}

// This will return all starred repos including the Atom feeds for their releases
func (gh *GitHubStarredFeedBuilder) GetStarredRepos() ([]GitHubRepo, error) {
	var allRepos []GitHubRepo
	url := "http://api.github.com/user/starred"

	log.Info("Querying GitHub for starred repos...")
	for {
		ghResponse, err := gh.doApiRequest(url)
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		err = json.Unmarshal(ghResponse.data, &repos)
		if err != nil {
			return nil, err
		}

		for i := range repos {
			repos[i].BuildAtomFeedUrl()
			allRepos = append(allRepos, repos[i])
		}

		// If there is no next page we are done...
		if ghResponse.nextPage == "" {
			return allRepos, nil
		}

		log.Debugf("Found next page: %s", ghResponse.nextPage)
		url = ghResponse.nextPage
	}
}

// This method handles making API requests to GitHub using its REST API.
func (gh *GitHubStarredFeedBuilder) doApiRequest(url string) (*GithubResponse, error) {
	logger := gh.logger

	headers := map[string]string{
		"Authorization":        fmt.Sprintf("Bearer %s", gh.token),
		"X-Github-Api-Version": "2022-11-28",
		"User-Agent":           "github.com/atomicmeganerd/gh-rel-to-rss",
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}

	httpRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Unable to build request to github", err)
		return nil, err
	}

	for k, v := range headers {
		httpRequest.Header.Set(k, v)
	}

	res, err := gh.client.Do(httpRequest)
	if err != nil {
		logger.Error("Unable to make request to Github", err)
		return nil, err

	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned an http error code %d", res.StatusCode)
	}

	ghResponse, err := gh.processGithubResponse(res)
	if err != nil {
		logger.Error("Unable to parse response from github", err)
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
