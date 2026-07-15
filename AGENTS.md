# Agent Guide for Starfeed

## Overview

Starfeed is a Go application that syncs release RSS feeds from support Git Forges on the Internet.
It adds feeds for new releases for starred repos to a specified FreshRSS instance. It uses very few
dependencies and is containerized via a multi-stage Dockerfile. CI builds, tests, lints, and
publishes images via GitHub Actions.

Currently supported:

- GitHub
- Forgejo based (including Codeberg)

## Config

Use the `STARFEED_CONFIG_PATH` environment variable to point Starfeed to the correct path for the
TOML file.

The TOML config contains secrets so must be kept in gitignore. On my server nix will deploy the TOML
securely.

```toml
debug=true
single_run=true

[[git_forges]]
type = "github"
name = "GitHub"
fqdn = "github.com"
token = "GITHUB_TOKEN"

[[git_forges]]
type = "forgejo"
name = "Codeberg"
fqdn = "codeberg.org"
token = "CODEBERG_TOKEN"

[rss_server]
name = "freshrss"
url = "http://freshrss:80"
user = "chris@megaparsec.ca"
token = "FRESHRSS_TOKEN"
```

**NOTE**: The TOML config must not be included in the docker image but mounted into the container.

## Tooling

- [Go Task](https://taskfile.dev)
- golangci-lint
- goimports
- golines
- gopls

I use logging for all debugging and never use a debugger.

## Code Organization

- `cmd/main.go`: Application entrypoint; sets up logging, reads config, handles signals, schedules
  24h ticker, and invokes the runners.
- `config/`: Loads configuration from the environment into our Go objects.
- `common/`: Common code like http handling, etc.
- `runners/`: Orchestration layer which executes workflows against the RSS server and the Git
  Forges.
- `gitforge/`: Implementation code for all supported git forges.
- `rss/`: Code that handles publishing the release feeds to RSS.
- `testutils/`: Test doubles related data and shared mocks/functions.

## Policies

- Interfaces for external I/O (HTTP clients) to enable mocking.
- Context passed to all network-bound components.
- Secret values (tokens) must never be logged.
- Use slog for structured logging.
- Line length limit 100 chars.
- For tests used shared mocks and data consts from `testutils/` package when possible.

## Testing

- `task test` uses `gotestsum` and generates `cover.out` for coverage.
