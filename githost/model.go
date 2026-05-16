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
	Kind     string
}

func (gr *BaseRepo) Name() string {
	return gr.RepoName
}

func (gr *BaseRepo) FeedURL() string {
	switch gr.Kind {
	case "forgejo":
		return fmt.Sprintf("%s/releases.rss", gr.RepoURL)
	case "github":
		return fmt.Sprintf("%s/releases.atom", gr.RepoURL)
	}
	// We validate this so we should never get here..
	return ""
}

// This is the response we get from the Git Host
type GitHostResponse struct {
	Data     []byte
	NextPage string
}
