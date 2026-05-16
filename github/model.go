package github

import (
	"fmt"
)

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type GitHubRepo struct {
	Name           string `json:"name"`
	HTMLURL        string `json:"html_url"`
	ReleaseFeedURL string
}

func (gr *GitHubRepo) String() string {
	return fmt.Sprintf("Name: %s, Releases Feed: %s", gr.Name, gr.ReleaseFeedURL)
}

func (gr *GitHubRepo) BuildReleasesFeedURL() {
	gr.ReleaseFeedURL = fmt.Sprintf("%s/releases.atom", gr.HTMLURL)
}

// This is the response we get from GitHub
type GitHubResponse struct {
	data     []byte
	nextPage string
}
