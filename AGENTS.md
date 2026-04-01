# Agent Guide for Starfeed

## Overview

Starfeed is a Go application that syncs GitHub release RSS feeds for repos you have starred into a FreshRSS instance. It uses only the Go standard library and is containerized via a multi-stage Dockerfile. CI builds, tests, lints, and publishes images via GitHub Actions.

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
- `STARFEED_GITHUB_API_TOKEN`
- `STARFEED_FRESHRSS_URL`
- `STARFEED_FRESHRSS_USER`
- `STARFEED_FRESHRSS_API_TOKEN`

Optional:
- `STARFEED_DEBUG_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_SINGLE_RUN_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_HTTP_TIMEOUT` (seconds; default 10)


**NOTE**: The app defaults to 24-hour intervals unless `STARFEED_SINGLE_RUN_MODE=true`.

Local dev: keep secrets in `.envrc` (direnv) and symlink `.env` -> `.envrc` for compose.

## Running in Containers

- Compose service uses: `docker-compose.yml`
  - Image set to `localhost/starfeed:latest`
  - `env_file: .envrc` (dotenv-style)
  - `restart: unless-stopped`
- Compose expects `.envrc` to be dotenv-style (KEY=VALUE); if using direnv functions,
  symlink a plain `.env`.

## Code Organization

- `cmd/main.go`: Application entrypoint; sets up logging, reads config, handles signals, schedules
  24h ticker, and invokes the publisher.
- `config/`: Configuration loading and validation from environment.
- `runner/`: Orchestrates querying GitHub, checking feeds, publishing to FreshRSS, and pruning
  stale feeds; uses goroutines and a WaitGroup.
- `github/`: GitHub API client (stars), pagination parsing, and release feed URL construction.
- `freshrss/`: FreshRSS client for authentication and feed management.
- `atom/`: Atom feed checker to ensure feeds have entries before adding.
- `mocks/`: Test doubles.

## Architecture

3 layers:

1. `cmd` — wiring and startup only. May log at any level.
2. `runner` — orchestration. May log at any level. The only layer that calls business logic.
3. `github`, `freshrss`, `atom` — business logic. No knowledge of upper layers. Debug logs only.

`config` is a utility package used by `cmd` for startup validation.

## Patterns and Conventions

- Interfaces for external I/O (HTTP clients) to enable mocking.
- Context passed to all network-bound components.
- Secret values (tokens) must never be logged.
- Use slog for structured logging; level toggled by `STARFEED_DEBUG_MODE`.
- Guard/short-circuit style to keep nesting shallow.
- Coverage threshold enforced in Taskfile (>=80%).
- Line length limit 100 chars.
- Ignore lints on resp.Body.Close() calls as that method never returns an error.

## Testing

- Unit tests present across packages: `*_test.go`.
- `task test` builds first, runs `go test` with `-race` and coverage, then checks threshold.
  Generates `cover.out` and `coverage.html` via `task generate-test-reports`.

## Dockerfile Notes

- Multi-stage build; builder uses latest go alpine image and installs `go-task`.
- Binary built to `/app/bin/starfeed`; CGO disabled and binary stripped (`-s -w`).
- Runner uses latest alpine image, non-root user created; `COPY --chown` and `PATH=/app/bin`.

## CI/CD

- `.github/workflows/release.yml` builds, lints, tests on push/PR.
- Publishes Docker images `atomicmeganerd/starfeed` with tags `latest` and version from `VERSION`.
- Tags repo and creates GitHub Release attaching `bin/starfeed`.
