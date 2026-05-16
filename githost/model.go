package githost

import "context"

type GitHost interface {
	GetStarredRepos(context.Context) (map[string]Repo, error)
	IsReleaseFeed(string) bool
}

type Repo interface {
	Name() string
	FeedURL() string
}

type GitHostType string

const (
	GitHub  GitHostType = "github"
	Forgejo GitHostType = "forgejo"
)

func (ght GitHostType) Valid() bool {
	switch ght {
	case GitHub, Forgejo:
		return true
	default:
	}
	return false
}

type GitHostConfig struct {
	Type    GitHostType
	BaseURL string
	Token   string
}
