# Changelog

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
