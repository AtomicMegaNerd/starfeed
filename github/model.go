package github

import (
	"fmt"
)

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type GitHubRepo struct {
	Name    string `json:"name"`
	HtmlUrl string `json:"html_url"`
	FeedUrl string
}

func (gr *GitHubRepo) String() string {
	return fmt.Sprintf("Name: %s, Releases Feed: %s", gr.Name, gr.FeedUrl)
}

func (gr *GitHubRepo) BuildReleasesFeedURL() {
	gr.FeedUrl = fmt.Sprintf("%s/releases.atom", gr.HtmlUrl)
}

// This is the response we get from Github
type GithubResponse struct {
	data     []byte
	nextPage string
}
