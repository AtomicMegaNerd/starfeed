# Starfeed

![Starfeed](./img/starfeed.png)

Starfeed scans the current list of your starred repos from any supported Git Forge on the Internet,
grabs the Releases RSS feed for each repo it finds, and publishes them to your own self-hosted
[FreshRSS](https://www.freshrss.org/) RSS aggregator. Then by hooking up an RSS client to your
FreshRSS server you can easily follow the releases for any of the repos that you have starred.

Starfeed will omit any RSS feeds that do not contain releases. It will also remove any feeds for
repos that you are no longer starring.

Currently supported Git Forges:

- GitHub
- Forgejo based (including Codeberg)

---

## Pre-Requisites

### Required Software

- You must have [FreshRSS](https://www.freshrss.org) deployed in your local network. It must be
  reachable from the Starfeed Docker container.
- You must have an API token generated in FreshRSS that has permissions to create/edit/delete feeds.
- You must have an API token for each Git Forge with permission to read starred repos.
- You must have [Docker](https://docker.com) or [Podman](https://podman.io) set up to run the
  container.
- To build and run the app locally you need to install [Go](https://go.dev),
  [Taskfile](https://taskfile.dev), and [Direnv](https://direnv.net/).

---

## Configuration

Starfeed uses a TOML configuration file. Set the `STARFEED_CONFIG_PATH` environment variable to
point to the file location. If unset, it defaults to `./starfeed.toml`.

### Example `starfeed.toml`

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

### Configuration Fields

| Field              | Description                                                        |
| ------------------ | ------------------------------------------------------------------ |
| `debug`            | Enable debug logging (`true`/`false`).                             |
| `single_run`       | Run once and exit (`true`) or run on a 24-hour interval (`false`). |
| `git_forges`       | List of Git Forge configurations. At least one is required.        |
| `git_forges.type`  | Forge type: `github` or `forgejo`.                                 |
| `git_forges.name`  | Display name for the forge.                                        |
| `git_forges.fqdn`  | Fully qualified domain name (e.g. `github.com`, `codeberg.org`).   |
| `git_forges.token` | API token with permission to read starred repos.                   |
| `rss_server.name`  | RSS server type: `freshrss`.                                       |
| `rss_server.url`   | URL of the FreshRSS instance.                                      |
| `rss_server.user`  | FreshRSS username/email.                                           |
| `rss_server.token` | FreshRSS API token.                                                |

**NOTE**: The TOML config contains secrets and must not be committed to version control or included
in Docker images. It should be mounted into the container as a volume.

---

## Setting the Environment

For local development, the best way to manage the environment is with [Direnv](https://direnv.net/).
Create an `.envrc` file (already in `.gitignore`).

### .envrc

```bash
use flake

source .env
```

### .env

The `.env` file is used by `docker-compose` to configure the local FreshRSS test harness. These
variables are **not** used by Starfeed itself.

```bash
# Used by docker-compose to provision the local FreshRSS test instance.
export FRESHRSS_USER=chris@megaparsec.ca
export FRESHRSS_PASS=*********
export FRESHRSS_API_TOKEN=*********
```

Then activate your environment with:

```bash
direnv allow
```

This will load all of the environment variables in `.envrc` into your environment while you are in
the project directory. See the direnv docs for more information.

## Running Locally with Containers (Recommended)

Use either **Docker** or **Podman** to run Starfeed. The instructions below show both options.

The included `docker-compose.yml` file will run FreshRSS and Starfeed locally. As long as the
environment is set up correctly above it will configure Starfeed to connect to this local test
instance of FreshRSS. Note that we use tmpfs so that the data is not persisted for FreshRSS after
the container is shut down.

While I use podman as my container runtime, `docker-compose` (the Go version) is pretty much a
requirement as `podman-compose` has bugs that break basics like `depends_on` with `healthcheck`.

### Using Docker Compose

If you want to build and run the app locally:

```bash
task docker-up
```

To stop the containers:

```bash
task docker-down
```

## Build and Run Go Binary (Not Recommended)

### Build

```bash
task build
```

### Run

As long as you have a valid `starfeed.toml` config file, you should be able to run the app. However,
you'll need to point to an existing FreshRSS instance. Generally the docker-compose option is
superior as it will spin up a test FreshRSS instance for you.

```bash
task run
```

### Test

To run the tests:

```bash
task test
```

### Lint

To lint the code:

```bash
task lint
```

## Cutting a Release

We use [GoReleaser](https://goreleaser.com/) with a GitHub action to create a new release of
Starfeed. This will publish our Docker image to Docker Hub.

Steps:

- Create a PR with your changes and ensure CI passes.
- Merge the PR to `main`.
- Go to the **Actions** tab on GitHub.
- Select the **Starfeed Release** workflow from the sidebar.
- Click **Run workflow** and enter the version tag in semver format (e.g. `v0.5.1`).
- The workflow will run CI first, then validate the version, create and push the git tag, and run
  GoReleaser to build binaries and publish the Docker image to Docker Hub.

## Roadmap

These are future items that I want to focus on.

### High Priority

I really want to get these items done.

- [x] Migrate to a config file (TOML) that Nix will securely deploy using agenix.
- [ ] Monitoring
- [ ] Setup integration test with a local test instance of Forgejo in the `docker-compose.yml` file.

### Backlog

These are less important and may or may not happen.

- [ ] Migrate to [codeberg](https://codeberg.org) or another host?
- [ ] Support other RSS backends?
- [ ] Support other Git Forges (Gitea, Bitbucket, Gitlab)?
- [ ] Create RSS feeds for notifications that we can serve up from this daemon.
- [ ] Specify watched instead of starred repos?
