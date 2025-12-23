# Agent Guide for Starfeed

## Overview

Starfeed is a Go application that syncs GitHub release RSS feeds for repos you have starred into a FreshRSS instance. It uses only the Go standard library and is containerized via a multi-stage Dockerfile. CI builds, tests, lints, and publishes images via GitHub Actions.

## Environment and Tooling

- Nix flake with nix-direnv is used for local development.
  - Enable the flake environment to get the Go toolchain: run `direnv allow` with `.envrc` containing `use flake`.
- Taskfile is used for build/test/lint commands.
- Podman or Docker can build and run the container. Podman is preferred locally.
- golangci-lint configuration enforces 100-char line length via `lll`.

## Essential Commands

- Build binary: `task build`
- Run locally: `task run`
- Test with coverage gate (>=80%): `task test`
- Lint: `task lint`
- Clean artifacts: `task clean`
- Update deps: `task update-deps`
- Generate coverage HTML: `task generate-test-reports`

## Environment Variables

Required (config.NewConfig enforces):
- `STARFEED_GITHUB_API_TOKEN`
- `STARFEED_FRESHRSS_URL`
- `STARFEED_FRESHRSS_USER`
- `STARFEED_FRESHRSS_API_TOKEN`

Optional:
- `STARFEED_DEBUG_MODE` ("true" for debug logs)
- `STARFEED_SINGLE_RUN_MODE` ("true" to exit after first run)
- `STARFEED_HTTP_TIMEOUT` (seconds; default 10)

Local dev: keep secrets in `.envrc` (direnv) and symlink `.env` -> `.envrc` for compose.

## Running in Containers

- Build image with explicit tag to avoid auto-names: `podman build -t localhost/starfeed:latest .`
- Compose service uses: `docker-compose.yml`
  - Image set to `localhost/starfeed:latest`
  - `env_file: .envrc` (dotenv-style)
  - `restart: unless-stopped`

## Code Organization

- `cmd/main.go`: Application entrypoint; sets up logging, reads config, handles signals, schedules 24h ticker, and invokes the publisher.
- `config/`: Configuration loading and validation from environment.
- `runner/`: Orchestrates querying GitHub, checking feeds, publishing to FreshRSS, and pruning stale feeds; uses goroutines and a WaitGroup.
- `github/`: GitHub API client (stars), pagination parsing, and release feed URL construction.
- `freshrss/`: FreshRSS client for authentication and feed management.
- `atom/`: Atom feed checker to ensure feeds have entries before adding.
- `mocks/`: Test doubles.

## Patterns and Conventions

- Interfaces for external I/O (HTTP clients) to enable mocking.
- Context passed to all network-bound components.
- Secret values (tokens) must never be logged.
- Use slog for structured logging; level toggled by `STARFEED_DEBUG_MODE`.
- Guard/short-circuit style to keep nesting shallow.
- Coverage threshold enforced in Taskfile (>=80%).
- Line length limit 100 chars (`.golangci.yml`).

## Testing

- Unit tests present across packages: `*_test.go` files in `atom/`, `github/`, `freshrss/`, `config/`, `runner/`.
- `task test` builds first, runs `go test` with `-race` and coverage, then checks threshold. Generates `cover.out` and `coverage.html` via `task generate-test-reports`.

## Dockerfile Notes

- Multi-stage build; builder uses `golang:1.25.5-alpine3.23` and installs `go-task`.
- Binary built to `/app/bin/starfeed`; CGO disabled and binary stripped (`-s -w`).
- Runner uses `alpine:3.23`, non-root user created; `COPY --chown` and `PATH=/app/bin`.

## CI/CD

- `.github/workflows/release.yml` builds, lints, tests on push/PR.
- Publishes Docker images `atomicmeganerd/starfeed` with tags `latest` and version from `VERSION`.
- Tags repo and creates GitHub Release attaching `bin/starfeed`.

## Gotchas

- Compose expects `.envrc` to be dotenv-style (KEY=VALUE); if using direnv functions, symlink a plain `.env`.
- Podman may auto-tag built images if `-t` not provided (e.g., `localhost/starfeed_starfeed`). Always tag explicitly.
- The app defaults to 24-hour intervals unless `STARFEED_SINGLE_RUN_MODE=true`.
- Ensure FreshRSS and GitHub tokens are valid; failures short-circuit publishing.
- Go 1.25 features used: `strings.SplitSeq` returns `iter.Seq[string]` and range permits only one iteration variable; use `for x := range strings.SplitSeq(...){}` (not `for _, x := range ...`). Do not replace with `strings.Split` unless necessary.

## Contribution Policies

- Keep code simple; prefer interfaces for DI.
- Always run `task build`, `task test`, `task lint` after changes.
- Maintain coverage >=80%.
- Do not modify `README.md` or `CLAUDE.md` via agents.
- Follow short-line (<100 chars) rule.
