package runner

import (
	"fmt"
	"net/http"

	"github.com/atomicmeganerd/starfeed/forgejo"
	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/github"
)

func GetGitHostFromConfig(
	gitHost githost.GitHostConfig,
	client *http.Client,
) (githost.GitHost, error) {
	switch gitHost.Type {
	case githost.GitHub:
		return github.NewGitHubStarredFeedBuilder(gitHost, client), nil
	case githost.Forgejo:
		return forgejo.NewForgejoStarredFeedBuilder(gitHost, client), nil
	}
	return nil, fmt.Errorf("invalid githost type %s", gitHost.Type)
}
