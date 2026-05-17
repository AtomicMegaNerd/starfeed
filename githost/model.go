package githost

import (
	"fmt"
)

type Repo interface {
	Name() string
	FeedURL() string
}

// This object represents a Git repo in a supported Git Host that is starred and that we want to
// get the Atom feed for.
type BaseRepo struct {
	RepoName string `json:"name"`
	RepoURL  string `json:"html_url"`
}

func (gr *BaseRepo) FeedURL() string {
	return fmt.Sprintf("%s/releases.atom", gr.RepoURL)
}

func (gr *BaseRepo) Name() string {
	return gr.RepoName
}

type ForgejoRepo struct {
	BaseRepo
	HasReleases bool `json:"has_releases"`
}

func (gr *ForgejoRepo) FeedURL() string {
	if !gr.HasReleases {
		return ""
	}
	return fmt.Sprintf("%s/releases.atom", gr.RepoURL)
}

// This is the response we get from the Git Host
type GitHostResponse struct {
	Data     []byte
	NextPage string
}
