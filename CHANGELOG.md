# Changelog

## [0.5.7](https://github.com/plapko/garminslacknotify/compare/v0.5.6...v0.5.7) (2026-04-27)


### Bug Fixes

* **garmin:** remove warmup that caused CSRF mismatch, add JWT_WEB Bearer auth ([d596542](https://github.com/plapko/garminslacknotify/commit/d596542d09c7bd43a3ed672ba6598090fcfa9b32))

## [0.5.6](https://github.com/plapko/garminslacknotify/compare/v0.5.5...v0.5.6) (2026-04-27)


### Bug Fixes

* **garmin:** add browser fingerprinting headers and use /sso/login for Spring warmup ([b9c37ef](https://github.com/plapko/garminslacknotify/commit/b9c37ef7b0c4e347f3f77f7172509f78f7832f9c))

## [0.5.5](https://github.com/plapko/garminslacknotify/compare/v0.5.4...v0.5.5) (2026-04-27)


### Bug Fixes

* **garmin:** authenticate with Spring via SSO ticket to obtain SESSIONID ([87367b4](https://github.com/plapko/garminslacknotify/commit/87367b45dffbca6aa4b994eb7ead4ce52a0c237d))

## [0.5.4](https://github.com/plapko/garminslacknotify/compare/v0.5.3...v0.5.4) (2026-04-27)


### Bug Fixes

* **garmin:** fetch specific /modern/css/ asset to warm up Spring session ([8a31ce2](https://github.com/plapko/garminslacknotify/commit/8a31ce26fab1c56baf7ca74152fd68d4a91fadc0))

## [0.5.3](https://github.com/plapko/garminslacknotify/compare/v0.5.2...v0.5.3) (2026-04-27)


### Bug Fixes

* warmup Spring session to obtain SESSIONID before /gc-api/ calls ([6f0c027](https://github.com/plapko/garminslacknotify/commit/6f0c0273363a1c9c24fc0e02960485fe68226b7f))

## [0.5.2](https://github.com/plapko/garminslacknotify/compare/v0.5.1...v0.5.2) (2026-04-27)


### Bug Fixes

* **garmin:** strip Origin/Referer from redirect requests in SSO flow ([7026356](https://github.com/plapko/garminslacknotify/commit/702635663bdb35536cbc6fd635ee11e379be4571))

## [0.5.1](https://github.com/plapko/garminslacknotify/compare/v0.5.0...v0.5.1) (2026-04-27)


### Bug Fixes

* **garmin:** remove extra headers, update User-Agent to Chrome 147/Windows ([9567cc0](https://github.com/plapko/garminslacknotify/commit/9567cc053c7a65f320af656eaa0483a77dbe5b2e))

## [0.5.0](https://github.com/plapko/garminslacknotify/compare/v0.4.9...v0.5.0) (2026-04-27)


### Features

* **garmin:** cache session cookies to avoid SSO on every cron run ([ecf79d8](https://github.com/plapko/garminslacknotify/commit/ecf79d86669eb03e1c3730a2b94a8b5907ffc5a4))

## [0.4.9](https://github.com/plapko/garminslacknotify/compare/v0.4.8...v0.4.9) (2026-04-27)


### Bug Fixes

* **garmin:** use service=/app/ and drop embed=true in SSO login ([f845aa3](https://github.com/plapko/garminslacknotify/commit/f845aa366c516e773e5aa907e540d96e137fb08a))

## [0.4.8](https://github.com/plapko/garminslacknotify/compare/v0.4.7...v0.4.8) (2026-04-27)


### Bug Fixes

* **garmin:** add Origin and Referer headers to /gc-api/ activities request ([876bb24](https://github.com/plapko/garminslacknotify/commit/876bb24d29cf05eec1b5866995524021931bce9b))

## [0.4.7](https://github.com/plapko/garminslacknotify/compare/v0.4.6...v0.4.7) (2026-04-27)


### Bug Fixes

* **garmin:** fetch /app/ explicitly if CSRF token missing after auth redirect ([1f1d190](https://github.com/plapko/garminslacknotify/commit/1f1d190f12898066d30427a0d5fc4a8c9148ddab))

## [0.4.6](https://github.com/plapko/garminslacknotify/compare/v0.4.5...v0.4.6) (2026-04-27)


### Bug Fixes

* **garmin:** return clear error on 429 rate limit instead of misleading 'check credentials' ([d381dd9](https://github.com/plapko/garminslacknotify/commit/d381dd92afba3e9e0af4af2bd45619bc70087f4e))

## [0.4.5](https://github.com/plapko/garminslacknotify/compare/v0.4.4...v0.4.5) (2026-04-27)


### Bug Fixes

* **garmin:** use /gc-api/ endpoint with connect-csrf-token header ([286db04](https://github.com/plapko/garminslacknotify/commit/286db04cba3f9b23cce9837e45b3c7c8e6f9d1fa))

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
