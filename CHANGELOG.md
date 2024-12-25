# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2024-12-24

### Fixed

- Added MIT license.

## [0.1.0] - 2024-12-24

### Added

- Added a cute logo. This is the first release ready for consumption.
- Added docker-compose file for local development.
- Updated documentation to reflect the new changes.

## [0.0.6] - 2024-12-24

### Added

- Added a github action to build and push the docker image to docker hub.

## [0.0.5] - 2024-12-24

### Added

- Added single run mode to sync feeds once and exit.

### Changed

- Migrated to log/slog for logging.
- Many small fixes for Docker and logging.

## [0.0.4] - 2024-12-23

### Added

- We only add to FreshRSS if the feed is not already there.
- Prune feeds from FreshRSS that are no longer starred.
- Run as a daemon once every 24 hours.

## [0.0.3] - 2024-12-19

### Added

- We now only publish a feed to FreshRSS if it has entries.
- We used a semaphore to throttle requests to FreshRSS so it does not get overwhelmed.

## [0.0.2] - 2024-12-18

### Added

- Publishing the Atom feeds from Github to FreshRSS works!

## [0.0.1] - 2024-12-12

### Added

- Initial release that is partially implemented.
- Can query GitHub API for the starred repos.
- Processes the response to get the list of Atom feeds for RSS.
