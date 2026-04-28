# garminslacknotify

Sets your Slack profile status with yesterday's Garmin Connect workout summary.

```
🏋️ Strength 45min & 🏊 Swimming 2.1km
```

Runs daily via crontab. Uses the unofficial Garmin Connect API.

## Install

**From GitHub Releases (recommended):**

Download the binary for your platform from the [Releases page](https://github.com/plapko/garminslacknotify/releases), extract it, and move it to `/usr/local/bin/`:

```bash
# Example for Linux amd64
curl -L https://github.com/plapko/garminslacknotify/releases/latest/download/garminslacknotify_Linux_x86_64.tar.gz | tar xz
sudo mv garminslacknotify /usr/local/bin/
```

**With Go:**

```bash
go install github.com/plapko/garminslacknotify/cmd/garminslacknotify@latest
```

## Setup

**1. First run — creates the config file automatically:**

```bash
garminslacknotify
# Config file created: ~/.config/garminslacknotify/config.yaml
# Edit it with your Garmin credentials and Slack token, then run again.
```

**2. Edit the config:**

```bash
nano ~/.config/garminslacknotify/config.yaml
```

Fill in:
- `garmin.email` / `garmin.password` — your Garmin Connect credentials
- `slack.token` — a Slack User OAuth token with `users.profile:write` scope

**Getting a Slack token:**

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → Create New App → From scratch
2. Name it (e.g. `garminslacknotify`), pick your workspace
3. Go to **OAuth & Permissions** → under **User Token Scopes** add: `users.profile:write`
4. Click **Install to Workspace** → copy the **User OAuth Token** (starts with `xoxp-`)

**3. Test it:**

```bash
garminslacknotify --dry-run
# Status text:  🏋️ Strength 45min & 🏊 Swimming 2.1km
# Status emoji: :muscle:
```

**4. Add to crontab (runs at 7:00 AM daily):**

```bash
crontab -e
```

Add this line:

```
0 7 * * * /usr/local/bin/garminslacknotify >> $HOME/garminslacknotify.log 2>&1
```

## Usage

```
Usage: garminslacknotify [flags]

Flags:
  --config    path to config file
              (default: ~/.config/garminslacknotify/config.yaml)
  --dry-run   print resulting status without setting it in Slack
  --date      override target date (YYYY-MM-DD, default: yesterday)
  --debug     print HTTP requests/responses and auth details to stderr
  --version   show version and exit
  --help      show this help

Examples:
  garminslacknotify                       run with default config
  garminslacknotify --dry-run             preview status, don't set it
  garminslacknotify --date 2026-04-20     check a specific past date
  garminslacknotify --config ~/my.yaml    use custom config path
  garminslacknotify --debug 2>debug.log   save debug output to file
```

## Activity Types Supported

The tool automatically maps Garmin activity types to emoji:

| Activity | Emoji |
|---|---|
| Running | 🏃 |
| Swimming | 🏊 |
| Strength Training | 🏋️ |
| Cycling | 🚴 |
| Yoga | 🧘 |
| Hiking | 🥾 |
| Walking | 🚶 |
| Rowing | 🚣 |
| Skiing | ⛷️ |
| Snowboarding | 🏂 |

Override any emoji in `config.yaml` under `activity_emojis`.

## Contributing

Commits follow [Conventional Commits](https://www.conventionalcommits.org/).
Releases are automated via [release-please](https://github.com/googleapis/release-please).

## License

MIT
