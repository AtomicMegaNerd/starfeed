package github

import (
	"fmt"
)

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type GitHubRepo struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	FullName         string `json:"full_name"`
	HtmlUrl          string `json:"html_url"`
	ReleasesAtomFeed string
}

func (gr *GitHubRepo) String() string {
	return fmt.Sprintf("Full Name: %s, Releases Feed: %s", gr.FullName, gr.ReleasesAtomFeed)
}

func (gr *GitHubRepo) BuildAtomFeedUrl() {
	gr.ReleasesAtomFeed = fmt.Sprintf("%s/releases.atom", gr.HtmlUrl)
}

type GithubResponse struct {
	data     []byte
	nextPage string
}
