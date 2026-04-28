# Changelog

## [1.0.2](https://github.com/plapko/garminslacknotify/compare/v1.0.1...v1.0.2) (2026-04-28)


### Bug Fixes

* normalize Garmin activity subtypes for emoji selection ([f2d4ca2](https://github.com/plapko/garminslacknotify/commit/f2d4ca2bb617092610076c611ec770e738796f3e))

## [1.0.1](https://github.com/plapko/garminslacknotify/compare/v1.0.0...v1.0.1) (2026-04-27)


### Bug Fixes

* create config directory with 0700 instead of 0755 ([be04078](https://github.com/plapko/garminslacknotify/commit/be04078dcfcaccbd9236a080998fe609d333b75d))

## [1.0.0](https://github.com/plapko/garminslacknotify/releases/tag/v1.0.0) (2026-04-27)

### Features

* Garmin Connect integration — SSO authentication with session caching (avoids re-login on every cron run)
* Slack profile status update with workout summary and activity emoji
* 10+ built-in activity type mappings (running, swimming, cycling, strength, yoga, hiking, …)
* Configurable emoji overrides per activity type via `config.yaml`
* Rest day status when no workouts are found
* `--dry-run` flag to preview the status without setting it
* `--date` flag to check a specific past date
* `--debug` flag for HTTP and auth troubleshooting
* First-run config file creation with annotated example
* Crontab-friendly — zero interaction after initial setup
