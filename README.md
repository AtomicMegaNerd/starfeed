# Starfeed

Starfeed scans the current list of your Github stars, grabs the Releases RSS feed for each repo it finds, and publishes them to your own self-hosted [FreshRSS](https://www.freshrss.org/) RSS aggregator. Then by hooking up an RSS client to your FreshRSS server you can easily follow the releases for any of the repos that you have starred.

Starfeed is written in Go. When done, Starfeed will be deployable as a Docker container in your home lab where it will happily update your GitHub RSS feeds on a interval.

## In Progress

So far we can get the atom feeds for any starred repositories. You need to set the following environment variables:

```bash
export GITHUB_USER=github_username
export GITHUB_API_TOKEN=github_token
export FRESHRSS_URL=url_to_freshrss
export FRESHRSS_USER=freshrss_user
export FRESHRSS_API_TOKEN=freshrss_api_token
```

The GitHub API token needs read starred repos access. The FreshRSS API token needs write access.

## Build and run

This app uses [Taskfile](https://taskfile.dev) to build and run the app. You can use the following command to build the app:

```bash
task build
```

To run the app:

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
- [ ] Only add feeds that are not already in FreshRSS
- [x] Only add a feed if it has entries
- [x] Come up with a better name
- [ ] Implement pruning of old feeds once they are no longer starred
- [ ] Containerize the app
- [ ] Make the app run on a schedule inside the container
- [ ] GitHub pipeline to build and publish the Docker image
- [ ] Write end-user documentation
- [ ] Add more tests
- [ ] Add some performance profiling
- [ ] Draw a cute logo in [Aseprite](https://www.aseprite.org/)
