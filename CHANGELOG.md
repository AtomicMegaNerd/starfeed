# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
