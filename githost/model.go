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

func (gr *BaseRepo) Name() string {
	return gr.RepoName
}

func (gr *BaseRepo) FeedURL() string {
	return fmt.Sprintf("%s/releases.atom", gr.RepoURL)
}

func (gr *BaseRepo) String() string {
	return fmt.Sprintf("Name: %s, Release Feed: %s", gr.Name(), gr.FeedURL())
}

// This is the response we get from the Git Host
type GitHostResponse struct {
	Data     []byte
	NextPage string
}
