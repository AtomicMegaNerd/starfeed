# Starfeed

![Starfeed](./img/starfeed.png)

Starfeed scans the current list of your Github stars, grabs the Releases RSS feed for each repo it finds, and publishes them to your own self-hosted [FreshRSS](https://www.freshrss.org/) RSS aggregator. Then by hooking up an RSS client to your FreshRSS server you can easily follow the releases for any of the repos that you have starred.

Starfeed will omit any RSS feeds from Github that are empty. It will also remove any feeds for repos that you are no longer starring.

Starfeed is written in Go and currently relies purely on the standard library with no external dependencies. The Docker image for this app is a little bigger than 25MB!

## Pre-Requisites

### Required Software

- You must have [FreshRSS](https://www.freshrss.org) deployed in your local network. It must be reachable from the Starfeed Docker container.
- You must have an API token generated in FreshRSS that has permissions to create/edit/delete feeds.
- You must have an API token for GitHub with permission to read starred repos.
- You must have [Docker](https://docker.com) or [Podman](https://podman.io) set up to run the container.
- To build and run the app locally you need to install [Go](https://go.dev), [Taskfile](https://taskfile.dev), and [Direnv](https://direnv.net/).

### Setting the Environment

The following environment variables need to be set for Starfeed to function correctly. For local
development, the best way is to create an `.envrc` file. This should remain in the `.gitignore` and
`.dockerignore` for obvious reasons.

```bash
use flake

export STARFEED_GITHUB_API_TOKEN=github_token
export STARFEED_FRESHRSS_URL=url_to_freshrss
export STARFEED_FRESHRSS_USER=freshrss_user
export STARFEED_FRESHRSS_API_TOKEN=freshrss_api_token
export STARFEED_DEBUG_MODE=true
export STARFEED_SINGLE_RUN_MODE=true
```

Then activate your environment with:

```bash
direnv allow
```

This will load all of the environment variables in `.envrc` into your environment while you are in the project directory. See the direnv docs for more information.

## Running with Docker or Podman

You can use either **Docker** or **Podman** to run Starfeed. The instructions below show both options.

### Using Docker

#### Pull and Run the Image

```bash
docker pull atomicmeganerd/starfeed:latest
docker run --env-file $PATH_TO_ENV_FILE -t atomicmeganerd/starfeed:latest
```

#### Using Docker Compose

If you want to build and run the app locally:

```bash
docker-compose up --build
```

### Using Podman

#### Pull and Run the Image

```bash
podman pull atomicmeganerd/starfeed:latest
podman run --env-file $PATH_TO_ENV_FILE -t atomicmeganerd/starfeed:latest
```

#### Using Podman Compose

If you want to build and run the app locally:

```bash
podman-compose up --build
```

---

### Build and Run Go Binary (Local Development)

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
- [x] GitHub pipeline to build and publish the Docker image
- [x] Write end-user documentation
- [x] Add some performance profiling
- [x] Draw a cute logo
- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Add test coverage to Taskfile and to Github Actions

## LLM Policies

### General Policies

- Always run `task build`, `task test`, and `task lint` after making any changes to the code.
- Always ensure each line is less than 100 characters long regardless of the file type.
- Only the human is allowed to make changes to the README.md file.

### Git Policies

- When the human asks you to make a commit, always create a new branch named `feature/short-description-of-change` and make the commit there.
- When the human asks you to commit, they are giving you explicit branch to make
  the commit

### Go Policies

- Keep the code as simple as possible.
- Use guards and short-circuiting to avoid deep nesting.
- Use custom error types if it makes the code simpler to read.
- Always handle errors explicitly.
- Always use context in functions that make network calls or do I/O.
- Always use dependency injection for anything that does I/O or network calls.
