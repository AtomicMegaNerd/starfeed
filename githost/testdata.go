package githost

import (
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/config"
)

var (
	MockValidGitHub = func(client *http.Client, logger *slog.Logger) GitHost {
		return NewGitHost(config.MockValidGitHubConfig, logger, client)
	}

	MockValidCodeberg = func(client *http.Client, logger *slog.Logger) GitHost {
		return NewGitHost(config.MockValidCodebergConfig, logger, client)
	}
)
