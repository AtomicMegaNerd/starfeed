# Starfeed

Starfeed scans the current list of your Github stars, grabs the Releases RSS feed for each repo it finds, and publishes them to your own self-hosted [FreshRSS](https://www.freshrss.org/) RSS aggregator. Then by hooking up an RSS client to your FreshRSS server you can easily follow the releases for any of the repos that you have starred.

Starfeed will omit any RSS feeds from Github that are empty. It will also remove any feeds for repos that you are no longer starring.

Starfeed is written in Go and currently relies purely on the standard library with no external dependencies. The Docker image for this app is a little bigger than 25MB!

## Building and Running Locally

### Pre-Requisites

- You must have [FreshRSS](https://www.freshrss.org) deployed in your local network. It much be reachable from the Starfeed Docker container.
- You must have [Docker](https://docker.com) or [Podman](https://podman.io) setup to run the container.
- To build and run the app locally you need to install [Go](https://go.dev), [Taskfile](https://taskfile.dev), and [Direnv](https://direnv.net/).

### Setting the Environment

The following environment variables need to be set for Starfeed to function correctly. For local
development the best way is to create an `.env` file. This should remain in the .gitigore and
.dockerignore for obvious reasons.

```bash
export STARFEED_GITHUB_USER=github_username
export STARFEED_GITHUB_API_TOKEN=github_token
export STARFEED_FRESHRSS_URL=url_to_freshrss
export STARFEED_FRESHRSS_USER=freshrss_user
export STARFEED_FRESHRSS_API_TOKEN=freshrss_api_token
export STARFEED_DEBUG_MODE=true
export STARFEED_SINGLE_RUN_MODE=true
```

To setup `direnv` you can create a file called `.envrc`

```bash
source .env
```

Then activate your environment with:

```bash
direnv allow
```

This will load all of the environment variables in `.env` into your environment while you are in the project directory. See the direnv docs for more information.

### Remote permissions

The GitHub API token needs read starred repos access. The FreshRSS API token needs write access.

### Build and run the Docker image

After setting the `.env` file above, you can build and run the container with the included shell script. From the root of the project run:

```bash
./scripts/run_docker.sh
```

### Build and run Go Binary

This app uses [Taskfile](https://taskfile.dev) to build and run the app. You can use the following command to build the app:

#### Build

```bash
task build
```

#### Run

As long as the environment variables are set up (with `direnv`) you should be able to run the app:

```bash
task run
```

To run the tests:

```bash
task test
```

## Tasks

- [x] Query Github for starred repos
- [x] Implement FreshRSS publishing
- [x] Add a semaphore to throttle the requests to FreshRSS
- [x] Only add feeds that are not already in FreshRSS
- [x] Only add a feed if it has entries
- [x] Come up with a better name
- [x] Implement pruning of old feeds once they are no longer starred
- [x] Containerize the app
- [x] Make the app run on a schedule inside the container
- [ ] GitHub pipeline to build and publish the Docker image
- [x] Write end-user documentation
- [ ] Add more tests
- [x] Add some performance profiling
- [ ] Draw a cute logo in [Aseprite](https://www.aseprite.org/)
