# Changelog

<!-- markdownlint-disable MD024 -->

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.3](https://github.com/memes/gce-metric/compare/v1.3.2...v1.3.3) (2024-10-11)


### Miscellaneous Chores

* Release 1.3.3 ([0896f88](https://github.com/memes/gce-metric/commit/0896f88fb7ff30ab78efaa609bb166433a55693b))

## [1.3.2](https://github.com/memes/gce-metric/compare/v1.3.1...v1.3.2) (2024-08-27)


### Bug Fixes

* require context for effective project ID ([6bd958e](https://github.com/memes/gce-metric/commit/6bd958e57bbdeb98c5e3fdf0f2bffde16b3f96dd))
* Update go.mod to require 1.23 ([f8416b6](https://github.com/memes/gce-metric/commit/f8416b6661d727d88fd56dbce47dd78dd3c9ff15))

## [1.3.1](https://github.com/memes/gce-metric/compare/v1.3.0...v1.3.1) (2024-04-15)


### Bug Fixes

* Remove unused context variable ([d578668](https://github.com/memes/gce-metric/commit/d57866896bc02b7fb6bab92d999c315bd2867267))

## [1.3.0](https://github.com/memes/gce-metric/compare/v1.2.3...v1.3.0) (2023-10-10)


### Features

* Add --json option to list command ([7fa0021](https://github.com/memes/gce-metric/commit/7fa002154a2f159ce5fa3a3e3a7a0e65dfa084f6))
* Add data command to retrieve TimeSeries data ([62fc930](https://github.com/memes/gce-metric/commit/62fc9304f63334a1232d2495e79f0e9780dae625))


### Bug Fixes

* Log to stderr, not stdout ([aa6f3f8](https://github.com/memes/gce-metric/commit/aa6f3f81acbc70d6a0792874b758d07afe0a0fc3))

## [1.2.3](https://github.com/memes/gce-metric/compare/v1.2.2...v1.2.3) (2023-03-24)


### Bug Fixes

* Fix broken cosign action in goreleaser ([856ec83](https://github.com/memes/gce-metric/commit/856ec830b20966f2740bd4b9dc25b534ab39a4f6))

## [1.2.2](https://github.com/memes/gce-metric/compare/v1.2.1...v1.2.2) (2023-03-24)


### Bug Fixes

* Release-please and GoReleaser updates ([9057933](https://github.com/memes/gce-metric/commit/9057933c0ed97ead340a5a7b90e1128b1535fc8c))

## [1.2.1](https://github.com/memes/gce-metric/compare/1.2.0-rc1...v1.2.1) (2023-03-24)


### Bug Fixes

* Add trace logging to pipeline ([8d18b28](https://github.com/memes/gce-metric/commit/8d18b28300de5fe8dcade55851a78f56e6dc4412))
* **cmd:** Address unused variable in list/delete ([8b25227](https://github.com/memes/gce-metric/commit/8b25227142e2f7a010db1314f0a674afc7c02e44))
* Improve the command line descriptions ([9fe5f9b](https://github.com/memes/gce-metric/commit/9fe5f9baccbcff58b3341320441f5f98b48b69b5))
* Log to stdout by default ([1ee8a29](https://github.com/memes/gce-metric/commit/1ee8a294d8c4272a5bc518c0647895459c690d21))
* Prefer NotifyContext for signal propagation ([5c7d43d](https://github.com/memes/gce-metric/commit/5c7d43d6c14182b26294f98bae633c4702283ebd))
* Update Go dependencies for March 23 2023 ([a5f0d1d](https://github.com/memes/gce-metric/commit/a5f0d1dca3eb97014b60a9d703cde740ace04ccb))


### Miscellaneous Chores

* release 1.2.1 ([ca235e2](https://github.com/memes/gce-metric/commit/ca235e21b788e8f039b4c4b51d04214ec0cdf283))

## [1.2.0-rc1] - 2022-08-05

First release candidate after refactoring application logic to separate metric generation,
processing, and commands, into separate packages.

### Added

- Release process generates a signed SBOM for all published binaries and containers.
- Autodetection for deployments on Compute Engine and GKE; generated metrics will
  have appropriate labels for those deployments. Can be overridden by flag.

### Changed

- Standardised on [cobra](https://github.com/spf13/cobra) and [viper](https://github.com/spf13/viper)
  for option processing and configuration.
  > NOTE: arguments now have two leading hyphens; e.g. `--verbose`.

## Removed

## [1.1.1] - 2020-08-16

Minor release that changes the GCP resource type to _generic_node_ if it does not detect Compute Engine metadata.

### Added

### Changed

### Removed

## [1.1.0] - 2020-08-11

Allow user to manage custom metrics that have been create; avoid reaching quotas.

### Added

- Sub-commands to list and delete custom metrics

### Changed

### Removed

## [1.0.0] - 2020-08-08

Initial release of project.

### Added

### Changed

### Removed

[1.2.0-rc1]: https://github.com/memes/gce-metric/compare/1.1.1...1.2.0-rc1
[1.1.1]: https://github.com/memes/gce-metric/compare/1.1.0...1.1.1
[1.1.0]: https://github.com/memes/gce-metric/compare/1.0.0...1.1.0
[1.0.0]: https://github.com/memes/gce-metric/releases/tag/1.0.0
