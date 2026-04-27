# Changelog

## [0.4.4](https://github.com/plapko/garminslacknotify/compare/v0.4.3...v0.4.4) (2026-04-27)


### Bug Fixes

* manually add JWT_FGP cookie to activities request ([351720b](https://github.com/plapko/garminslacknotify/commit/351720b5245bd1596a5a32bd738612db991c7e29))

## [0.4.3](https://github.com/plapko/garminslacknotify/compare/v0.4.2...v0.4.3) (2026-04-27)


### Bug Fixes

* extract app CSRF token and send with activities request ([b81cd74](https://github.com/plapko/garminslacknotify/commit/b81cd746b7c943b2b5dceb8243b9fcd895fc6e1c))

## [0.4.2](https://github.com/plapko/garminslacknotify/compare/v0.4.1...v0.4.2) (2026-04-27)


### Bug Fixes

* handle {} response from Garmin activities endpoint ([2a0f573](https://github.com/plapko/garminslacknotify/commit/2a0f573606840499286666b6fa4d03a22ab6bce6))

## [0.4.1](https://github.com/plapko/garminslacknotify/compare/v0.4.0...v0.4.1) (2026-04-27)


### Bug Fixes

* use /proxy/ prefix for Garmin activities endpoint ([09aa418](https://github.com/plapko/garminslacknotify/commit/09aa418e772806bba48bce8a5825ed0585c027a1))

## [0.4.0](https://github.com/plapko/garminslacknotify/compare/v0.3.1...v0.4.0) (2026-04-27)


### Features

* add --debug flag for HTTP and auth troubleshooting ([f36978d](https://github.com/plapko/garminslacknotify/commit/f36978d36f884f6c5825207e0f2cb81495e02879))

## [0.3.1](https://github.com/plapko/garminslacknotify/compare/v0.3.0...v0.3.1) (2026-04-27)


### Bug Fixes

* capture Garmin SSO ticket from redirect URL ([f0f3794](https://github.com/plapko/garminslacknotify/commit/f0f37949c1907c649e5b498bdbfa0bcf028c56b5))
* extract PR number from JSON in auto-merge step ([f0f3794](https://github.com/plapko/garminslacknotify/commit/f0f37949c1907c649e5b498bdbfa0bcf028c56b5))

## [0.3.0](https://github.com/plapko/garminslacknotify/compare/v0.2.3...v0.3.0) (2026-04-27)


### Features

* add rich terminal UI with spinners and colored output ([a5e11a3](https://github.com/plapko/garminslacknotify/commit/a5e11a36442d2c9b4237b5a3a0274e64d89e1a7a))


### Bug Fixes

* add XMLHttpRequest header and improve activities error message ([a924938](https://github.com/plapko/garminslacknotify/commit/a924938f54f9ec0a70d5ffaf59be9c93fb66a86e))

## [0.2.3](https://github.com/plapko/garminslacknotify/compare/v0.2.2...v0.2.3) (2026-04-27)


### Bug Fixes

* add browser headers to Garmin SSO requests ([3839a1b](https://github.com/plapko/garminslacknotify/commit/3839a1baf78442e3b20dbcb1f7d71f35de5f37c4))

## [0.2.2](https://github.com/plapko/garminslacknotify/compare/v0.2.1...v0.2.2) (2026-04-27)


### Bug Fixes

* use PAT for release-please to allow GoReleaser workflow trigger ([38298c3](https://github.com/plapko/garminslacknotify/commit/38298c36e99b853ba7e91a96e6b7fb106c97a7f0))

## [0.2.1](https://github.com/plapko/garminslacknotify/compare/v0.2.0...v0.2.1) (2026-04-27)


### Bug Fixes

* trigger GoReleaser on release created event instead of tag push ([fdccdb9](https://github.com/plapko/garminslacknotify/commit/fdccdb9cf8f0a78a12711923ae71c54740afb0bd))

## [0.2.0](https://github.com/plapko/garminslacknotify/compare/v0.1.0...v0.2.0) (2026-04-27)


### Features

* add CLI entry point with flag parsing and first-run config ([66deece](https://github.com/plapko/garminslacknotify/commit/66deece8f7d3114a117e23c346e9edc9b01584f3))
* add config package with YAML load, validate, and template ([08b8b44](https://github.com/plapko/garminslacknotify/commit/08b8b44a9c8f0bce8d215d41552cdd5107672a37))
* add formatter package with emoji map and smart truncation ([140ee0a](https://github.com/plapko/garminslacknotify/commit/140ee0a717049770178980ad04651a0c6e08e82e))
* add Garmin Connect client with SSO auth and activity fetch ([6f9548c](https://github.com/plapko/garminslacknotify/commit/6f9548c38b3bc4d3c7f593cb131985a732f88cab))
* add Slack client for profile status ([a9e2089](https://github.com/plapko/garminslacknotify/commit/a9e20892d1b8d8f29158d781471c6311f6bc8935))


### Bug Fixes

* config validate preserves explicit rest_day.enabled setting ([a781f3e](https://github.com/plapko/garminslacknotify/commit/a781f3e615d54045156b4e238ef67f993ac56cff))
