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

## Environment Variables

- `STARFEED_GIT_FORGE_*n*` where _n_ is a number from 0..n. This is a CSV value with the following
  format: `type,name,fqdn,token`.
- `STARFEED_RSS_SERVER` which again uses CSV to configure our RSS server. Format:
  `type,url,user,token`.

### Optional

- `STARFEED_DEBUG_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_SINGLE_RUN_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_HTTP_TIMEOUT` (seconds; default 60)

**NOTE**: The app defaults to 24-hour intervals unless `STARFEED_SINGLE_RUN_MODE=true`.

### FreshRSS Test Harness for Docker Compose

These 3 variables can be used to configure a local FreshRSS instance running with docker-compose. If
you are not running this test harness locally these can be ignored.

- `STARFEED_RSS_USER` set this to the same user in `STARFEED_RSS_SERVER`.
- `STARFEED_RSS_API_TOKEN` this will be the same token you use in `STARFEED_RSS_SERVER`.
- `STARFEED_RSS_PASS` this is a separate password that you use to login to the GUI.

See the local `docker-compose.yml` to see how these variables are used.

### Setting the Environment

The following environment variables need to be set for Starfeed to function correctly. For local
development, the best way is to create an `.envrc` file. This should remain in the `.gitignore` and
`.dockerignore` for obvious reasons.

#### .envrc

```bash
use flake

source .env
```

#### .env

Below is an example file that shows how to configure the environment variables.

```bash
# Git Forges
export STARFEED_GIT_FORGE_0=forgejo,Codeberg,codeberg.org,*****************
export STARFEED_GIT_FORGE_1=github,GitHub,github.com,***************

# RSS Server
export STARFEED_RSS_SERVER=freshrss,http://freshrss:80,chris@megaparsec.ca,*****************

# Use these with `docker-compose.yml` if you want the FreshRSS test harness locally.
export STARFEED_RSS_USER=chris@megaparsec.ca
export STARFEED_RSS_PASS=*********
export STARFEED_RSS_API_TOKEN=*********

# Flags
export STARFEED_DEBUG_MODE=true
export STARFEED_SINGLE_RUN_MODE=true
```

Then activate your environment with:

```bash
direnv allow
```

This will load all of the environment variables in `.envrc` into your environment while you are in
the project directory. See the direnv docs for more information.

## Running with Docker or Podman

You can use either **Docker** or **Podman** to run Starfeed. The instructions below show both
options.

### Running Locally with Containers (Recommended)

The included `docker-compose.yml` file will run freshrss and starfeed locally. As long as the
environment is setup correctly above it will configure starfeed to connect to this local test
instance of FreshRSS. Note that we use tmpfs so that the data is not persisted for FreshRSS after
the container is shut down.

While I use podman as my container runtime, `docker-compose` (the Go version) is pretty much a
requirement as `podman-compose` has bugs that break basics like `depends_on` with `healthcheck`.

#### Using Docker Compose

If you want to build and run the app locally:

```bash
docker-compose up --build
```

### Build and Run Go Binary (Local Development)

This app uses [Taskfile](https://taskfile.dev) to build and run the app. You can use the following
command to build the app:

#### Build

```bash
task build
```

#### Run

As long as the environment variables are set up (with `direnv`) you should be able to run the app.
However, you'll need to point to an existing FreshRSS instance. Generally the docker/podman-compose
options is superior as it will spin up a test FreshRSS instance for you.

```bash
task run
```

To run the tests:

```bash
task test
```

## Roadmap

These are future items that I want to focus on.

### High Priority

I really want to get these items done.

- [ ] Monitoring
- [ ] Setup integration test with a local test instance of Forgejo in the `docker-compose.yml` file.
- [ ] Migrate to a config file (TOML) with secrets being stored in the environment.
- [ ] Nix flake for deploying to my NixOS server (and not just for setting up dev shell).

### Backlog

These are less important and may or may not happen.

- [ ] Migrate to [codeberg](https://codeberg.org) or another host?
- [ ] Support other RSS backends?
- [ ] Support other Git Forges (Gitea, Bitbucket, Gitlab)?
- [ ] Create RSS feeds for notifications that we can serve up from this daemon.
- [ ] Specify watched instead of starred repos?
