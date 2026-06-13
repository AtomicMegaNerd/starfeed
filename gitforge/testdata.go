package gitforge

import (
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/config"
)

var (
	MockValidGitHub = func(client *http.Client, logger *slog.Logger) GitForge {
		return NewGitForge(config.MockValidGitHubConfig, logger, client)
	}

	MockValidCodeberg = func(client *http.Client, logger *slog.Logger) GitForge {
		return NewGitForge(config.MockValidCodebergConfig, logger, client)
	}
)
