# Starfeed

![Starfeed](./img/starfeed.png)

Starfeed scans the current list of your starred repos from any supported Git Host on the Internet,
grabs the Releases RSS feed for each repo it finds, and publishes them to your own self-hosted
[FreshRSS](https://www.freshrss.org/) RSS aggregator. Then by hooking up an RSS client to your
FreshRSS server you can easily follow the releases for any of the repos that you have starred.

Starfeed will omit any RSS feeds that do not contain releases. It will also remove any feeds for
repos that you are no longer starring.

Currently supported Git Hosts:

- GitHub
- Forgejo based (including Codeberg)

## Pre-Requisites

### Required Software

- You must have [FreshRSS](https://www.freshrss.org) deployed in your local network. It must be
  reachable from the Starfeed Docker container.
- You must have an API token generated in FreshRSS that has permissions to create/edit/delete feeds.
- You must have an API token for each Git Host with permission to read starred repos.
- You must have [Docker](https://docker.com) or [Podman](https://podman.io) set up to run the
  container.
- To build and run the app locally you need to install [Go](https://go.dev),
  [Taskfile](https://taskfile.dev), and [Direnv](https://direnv.net/).

## Environment Variables

- `STARFEED_GIT_HOST_*n*` where _n_ is a number from 0..n. This is a CSV value with the following
  format: `type,name,url,api_url,token,enabled`.
- `STARFEED_RSS_SERVER` which again uses CSV to configure our RSS server. Format:
  `type,url,user,token,enabled`.

### Optional

- `STARFEED_DEBUG_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_SINGLE_RUN_MODE` (any value valid for strconv.ParseBool is good here)
- `STARFEED_HTTP_TIMEOUT` (seconds; default 60)

**NOTE**: The app defaults to 24-hour intervals unless `STARFEED_SINGLE_RUN_MODE=true`.

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
# Git Hosts
export STARFEED_GIT_HOST_0=forgejo,Codeberg,https://codeberg.org,https://codeberg.org/api/v1,*****************,true
export STARFEED_GIT_HOST_1=github,GitHub,https://github.com,https://api.github.com,***************,true

# RSS Server
export STARFEED_RSS_SERVER=freshrss,http://freshrss:80,chris@megaparsec.ca,*****************,true

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

The following `docker-compose.yml` file will run freshrss and starfeed locally. As long as the
environment is setup correctly above it will configure starfeed to connect to this local test
instance of FreshRSS. Note that we use tmpfs so that the data is not persisted for FreshRSS after
the container is shut down.

**NOTE** See
[https://github.com/containers/podman-compose/issues/1422](https://github.com/containers/podman-compose/issues/1422)
to see why we can't use depends_on with docker-compose right now.

```yaml
---
services:
  freshrss:
    image: freshrss/freshrss:latest
    container_name: freshrss-test
    restart: unless-stopped
    env_file:
      - .env
    ports:
      - "8080:80"
    tmpfs:
      - /var/www/FreshRSS/data
    environment:
      TZ: UTC
      FRESHRSS_INSTALL: |
        --api-enabled
        --base-url http://localhost:8080
        --default-user ${STARFEED_RSS_USER}
      FRESHRSS_USER: |
        --user ${STARFEED_RSS_USER}
        --password ${STARFEED_RSS_PASS}
        --email test@example.net
        --api-password ${STARFEED_RSS_API_TOKEN}
        --language en
  starfeed:
    image: localhost/starfeed:latest
    build:
      context: .
      dockerfile: Dockerfile
    command: ["sh", "-c", "sleep 5 && starfeed"]
    env_file:
      - .env
```

#### Using Docker Compose

If you want to build and run the app locally:

```bash
docker-compose up --build
```

#### Using Podman Compose

If you want to build and run the app locally:

```bash
podman-compose up --build
```

---

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
