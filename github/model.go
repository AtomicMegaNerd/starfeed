package github

import (
	"fmt"
)

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type GitHubRepo struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	HtmlUrl         string `json:"html_url"`
	ReleasesFeedUrl string
}

func (gr *GitHubRepo) String() string {
	return fmt.Sprintf("Full Name: %s, Releases Feed: %s", gr.FullName, gr.ReleasesFeedUrl)
}

func (gr *GitHubRepo) BuildReleasesFeedURL() {
	gr.ReleasesFeedUrl = fmt.Sprintf("%s/releases.atom", gr.HtmlUrl)
}

type GithubResponse struct {
	data     []byte
	nextPage string
}
