# gh-rel-to-rss

This program scans the current list of Github stars and creates an RSS feed for the releases
of the starred repositories. It will also prune any RSS feeds that are no longer starred.

## In Progress

So far we can get the atom feeds for any starred repositories. You need to set the following
environment variables:

```bash
export GITHUB_USER=github_username
export GITHUB_API_TOKEN=github_token
export FRESHRSS_URL=url_to_freshrss
export FRESHRSS_USER=freshrss_user
export FRESHRSS_API_TOKEN=freshrss_api_token
```

The GitHub API token needs read starred repos access. The FreshRSS API token needs write access.

## Build and run

This app uses Taskfile to build and run the app. You can use the following command to build the app:

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
- [ ] Add a semaphore to throttle the requests to FreshRSS
- [ ] Only add feeds that are not already in FreshRSS
- [ ] Only add a feed if it has entries
- [ ] Implement pruning of old feeds once they are no longer starred
- [ ] Dockerize the app
- [ ] Make the app run on a schedule inside the container
- [ ] GitHub pipeline to build and publish the Docker image
- [ ] Finish documentation
- [ ] Add more tests
