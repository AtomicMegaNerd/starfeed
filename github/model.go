package github

import (
	"fmt"
	"time"
)

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type GitHubRepo struct {
	Name           string `json:"name"`
	HtmlUrl        string `json:"html_url"`
	ReleaseFeedUrl string
	IssuesFeedUrl  string
}

func (gr *GitHubRepo) String() string {
	return fmt.Sprintf("Name: %s, Releases Feed: %s", gr.Name, gr.ReleaseFeedUrl)
}

func (gr *GitHubRepo) BuildReleasesFeedURL() {
	gr.ReleaseFeedUrl = fmt.Sprintf("%s/releases.atom", gr.HtmlUrl)
}

// This is the response we get from GitHub
type GitHubResponse struct {
	data     []byte
	nextPage string
}

type GitHubIssueBase struct {
	ID            int64      `json:"id"`
	Number        int64      `json:"number"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	HTMLURL       string     `json:"html_url"`
	UpdatedAt     time.Time  `json:"updated_at"`
	CreatedAt     time.Time  `json:"created_at"`
	User          GitHubUser `json:"user"`
	RepositoryURL string     `json:"repository_url"`
}

type GitHubIssue struct {
	GitHubIssueBase
	Labels GitHubIssueLabel `json:"labels"`
}

type GitHubIssueLabel struct {
	Name string `json:"name"`
}

type GitHubPullRequest struct {
	GitHubIssueBase
	PullRequest GitHubPullRequestField `json:"pull_request"`
}

type GitHubUser struct {
	Name string `json:"name"`
}

type GitHubPullRequestField struct {
	URL     string `json:"url"`
	HTMLURL string `json:"html_url"`
}

type GitHubIssueComment struct {
	ID        int64      `json:"id"`
	HTMLURL   string     `json:"html_url"`
	User      GitHubUser `json:"user"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedAt time.Time  `json:"created_at"`
}
