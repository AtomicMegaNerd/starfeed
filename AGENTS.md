# Agent Guide for Starfeed

## Overview

Starfeed is a Go application that syncs GitHub release RSS feeds for repos you have starred into a
FreshRSS instance. It uses very few dependencies and is containerized via a multi-stage Dockerfile.
CI builds, tests, lints, and publishes images via GitHub Actions.

## Environment and Tooling

- Nix flake with nix-direnv is used for local development.
  - Enable the flake environment to get the Go toolchain: run `direnv allow` with `.envrc`
    containing `use flake`.
- Taskfile is used for build/test/lint commands. Always use the task command to build, lint, test.
- Podman or Docker can build and run the container. Podman is preferred locally.

## Essential Commands

- Build binary: `task build`
- Run locally: `task run`
- Test with coverage gate (>=80%): `task test`
- Lint: `task lint`
- Clean artifacts: `task clean`
- Update deps: `task update-deps`
- Generate coverage HTML: `task generate-test-reports`
- Always run `task build`, `task test`, `task lint` after changes.

## Environment Variables

Required (config.NewConfig enforces):

- `STARFEED_GIT_HOST_*n*` where _n_ is a number from 0..n. This is a CSV value with the following
  format: `type,name,url,token`.
- `STARFEED_RSS_SERVER` which again uses CSV to configure our RSS server. Format:
  `type,url,user,token`.

Optional:

- `STARFEED_DEBUG_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_SINGLE_RUN_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_HTTP_TIMEOUT` (seconds; default 10)

**NOTE**: The app defaults to 24-hour intervals unless `STARFEED_SINGLE_RUN_MODE=true`.

Local dev: keep secrets in `.envrc` (direnv) and symlink `.env` -> `.envrc` for compose.

## Running in Containers

- Compose service uses: `docker-compose.yml`
  - Image set to `localhost/starfeed:latest`
  - `env_file: .env` (dotenv-style)
  - `restart: unless-stopped`

## Code Organization

- `cmd/main.go`: Application entrypoint; sets up logging, reads config, handles signals, schedules
  24h ticker, and invokes the runners.
- `config/`: Loads configuration from the environment into our Go objects.
- `runner/`: Orchestration layer which executes workflows based on the type of githost. feeds.
- `githost/`: Shared code for all supported gihosts.
- `github/`: Code that is specific to the GitHub kind of git host.
- `forgejo/`: Code that is specific to the Forgejo kind of git host.
- `rss/`: Code that handles publishing the release feeds to RSS.
- `atom/`: Atom feed checker to ensure feeds have entries before adding.
- `mocks/`: Test doubles related data and shared mocks/functions.

## Patterns and Conventions

- Interfaces for external I/O (HTTP clients) to enable mocking.
- Context passed to all network-bound components.
- Secret values (tokens) must never be logged.
- Use slog for structured logging; level toggled by `STARFEED_DEBUG_MODE`.
- Guard/short-circuit style to keep nesting shallow.
- Line length limit 100 chars.
- Ignore lints on resp.Body.Close() calls as that method never returns an error.
- For tests used shared mocks and data consts from `mocks/` package when possible.

## Testing

- `task test` uses `gotestsum` and generates `cover.out` for coverage.

## Dockerfile Notes

- Multi-stage build; builder uses latest go alpine image version (but not latest tag) and installs
  `go-task`.
- Binary built to `/app/bin/starfeed`; CGO disabled and binary stripped (`-s -w`).
- Runner uses latest alpine image version (not latest tag), non-root user created; `COPY --chown`
  and `PATH=/app/bin`.

## CI/CD

- `.github/workflows/release.yml` builds, lints, tests on push/PR.
- Publishes Docker images `atomicmeganerd/starfeed` with tags `latest` and version from `VERSION`.
- Tags repo and creates GitHub Release attaching `bin/starfeed`.
