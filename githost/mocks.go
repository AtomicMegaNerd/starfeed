package githost

import "github.com/atomicmeganerd/starfeed/mocks"

var MockValidGitHost = GitHostConfig{
	Type:    mocks.GitHubType,
	Name:    mocks.GitHubName,
	BaseURL: mocks.GitHubURL,
	Token:   mocks.GitHubToken,
}
