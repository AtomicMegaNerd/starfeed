package githost

import "github.com/atomicmeganerd/starfeed/mocks"

var (
	MockValidGitHub = GitHostConfig{
		Type:    mocks.GitHubType,
		Name:    mocks.GitHubName,
		BaseURL: mocks.GitHubURL,
		Token:   mocks.GitHubToken,
	}

	MockValidForgejo = GitHostConfig{
		Type:    mocks.ForgejoType,
		Name:    mocks.ForgejoName,
		BaseURL: mocks.ForgejoURL,
		Token:   mocks.ForgejoToken,
	}
)
