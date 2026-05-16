package github

import (
	"fmt"
)

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type GitHubRepo struct {
	RepoName string `json:"name"`
	HTMLURL  string `json:"html_url"`
}

func (gr *GitHubRepo) Name() string {
	return gr.RepoName
}

func (gr *GitHubRepo) FeedURL() string {
	return fmt.Sprintf("%s/releases.atom", gr.HTMLURL)
}

func (gr *GitHubRepo) String() string {
	return fmt.Sprintf("Name: %s, Release Feed: %s", gr.Name(), gr.FeedURL())
}

// This is the response we get from GitHub
type GitHubResponse struct {
	data     []byte
	nextPage string
}
