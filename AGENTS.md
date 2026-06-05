# Agent Guide for Starfeed

## Overview

Starfeed is a Go application that syncs release RSS feeds from support Git Hosts on the Internet. It
adds feeds for new releases for starred repos to a specified FreshRSS instance. It uses very few
dependencies and is containerized via a multi-stage Dockerfile. CI builds, tests, lints, and
publishes images via GitHub Actions.

Currently supported:

- GitHub
- Forgejo based (including Codeberg)

## Environment and Tooling

- Nix flake with nix-direnv is used for local development.
  - Enable the flake environment to get the Go toolchain: run `direnv allow` with `.envrc`
    containing `use flake`.
- Taskfile is used for build/test/lint commands. Always use the task command to build, lint, test.
- Podman or Docker can build and run the container. Podman is preferred locally.

## Environment Variables

- `STARFEED_GIT_HOST_*n*` where _n_ is a number from 0..n. This is a CSV value with the following
  format: `type,name,url,api_url,token,enabled`.
- `STARFEED_RSS_SERVER` which again uses CSV to configure our RSS server. Format:
  `type,url,user,token,enabled`.

Optional:

- `STARFEED_DEBUG_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_SINGLE_RUN_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_HTTP_TIMEOUT` (seconds; default 60)

**NOTE**: The app defaults to 24-hour intervals unless `STARFEED_SINGLE_RUN_MODE=true`.

## Essential Commands

- Build binary: `task build`
- Run locally: `task run`
- Test with coverage gate (>=80%): `task test`
- Lint: `task lint`
- Clean artifacts: `task clean`
- Update deps: `task update-deps`
- Generate coverage HTML: `task generate-test-reports`
- Always run `task build`, `task test`, `task lint` after changes.

## Code Organization

- `cmd/main.go`: Application entrypoint; sets up logging, reads config, handles signals, schedules
  24h ticker, and invokes the runners.
- `config/`: Loads configuration from the environment into our Go objects.
- `runner/`: Orchestration layer which executes workflows against the RSS server and the Git Hosts.
- `githost/`: Implementation code for all supported git hosts.
- `rss/`: Code that handles publishing the release feeds to RSS.
- `testutils/`: Test doubles related data and shared mocks/functions.

## Patterns and Conventions

- Interfaces for external I/O (HTTP clients) to enable mocking.
- Context passed to all network-bound components.
- Secret values (tokens) must never be logged.
- Use slog for structured logging; level toggled by `STARFEED_DEBUG_MODE`.
- Line length limit 100 chars.
- Ignore lints on resp.Body.Close() calls as that method never returns an error.
- For tests used shared mocks and data consts from `testutils/` package when possible.

## Testing

- `task test` uses `gotestsum` and generates `cover.out` for coverage.
