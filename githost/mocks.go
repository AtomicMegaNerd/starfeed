package githost

import (
	"net/http"

	"github.com/atomicmeganerd/starfeed/mocks"
)

var (
	MockValidGitHub = func(client *http.Client) GitHost {
		return &gitHost{
			hostType: mocks.GitHubType,
			name:     mocks.GitHubName,
			baseURL:  mocks.GitHubURL,
			token:    mocks.GitHubToken,
			client:   client,
		}
	}

	MockValidForgejo = func(client *http.Client) GitHost {
		return &gitHost{
			hostType: mocks.ForgejoType,
			name:     mocks.ForgejoName,
			baseURL:  mocks.ForgejoURL,
			token:    mocks.ForgejoToken,
			client:   client,
		}
	}
)
