# garminslacknotify — Design Spec

**Date:** 2026-04-27  
**Status:** Approved

## Overview

A Go CLI tool that runs daily via crontab, fetches yesterday's Garmin Connect workout data, and sets the user's Slack profile status with emoji-rich workout summary.

---

## Architecture

Layered architecture with `cmd/` entry point and independent `internal/` packages.

```
garminslacknotify/
├── cmd/garminslacknotify/
│   └── main.go              # CLI entry point, flag parsing
├── internal/
│   ├── config/
│   │   └── config.go        # YAML loading, validation, template generation
│   ├── garmin/
│   │   └── client.go        # Garmin Connect HTTP client (SSO auth + activities)
│   ├── slack/
│   │   └── client.go        # Slack users.profile.set API
│   └── formatter/
│       └── formatter.go     # Emoji mapping + smart truncation logic
├── .github/
│   └── workflows/
│       ├── ci.yml           # Build + test on every push
│       └── release.yml      # GoReleaser on tag push
├── .goreleaser.yaml         # GoReleaser config
├── .release-please-manifest.json
├── release-please-config.json
├── config.example.yaml      # Annotated example config
├── go.mod
├── go.sum
└── README.md
```

---

## CLI

```
Usage: garminslacknotify [flags]

Flags:
  --config    path to config file
              (default: ~/.config/garminslacknotify/config.yaml)
  --dry-run   print resulting status without setting it in Slack
  --date      override target date (YYYY-MM-DD, default: yesterday)
  --version   show version and exit
  --help      show this help

Examples:
  garminslacknotify                      # run with default config
  garminslacknotify --dry-run            # preview status, don't set it
  garminslacknotify --date 2026-04-20    # check a specific past date
  garminslacknotify --config ~/my.yaml   # use custom config path
```

### First-run behaviour

If the config file does not exist, the tool automatically:
1. Creates `~/.config/garminslacknotify/config.yaml` from an annotated template
2. Prints a message explaining where to get credentials
3. Exits with code 0

```
Config file created: ~/.config/garminslacknotify/config.yaml
Edit it with your Garmin credentials and Slack token, then run again.
```

If config exists but is invalid or has empty required fields, the tool exits with a descriptive error:
```
Error: garmin.email is required
```

---

## Config (YAML)

```yaml
garmin:
  email: "user@example.com"
  password: "your-password"

slack:
  # User OAuth token — needs users.profile:write scope
  # Get it at: https://api.slack.com/apps → your app → OAuth & Permissions
  token: "xoxp-..."

rest_day:
  enabled: true       # set status when no workouts found (default: true)
  text: "Rest day"
  emoji: "sleeping"   # Slack emoji name without colons

# Optional: override default emoji per Garmin activity type key
# Built-in defaults cover ~20 common activity types
activity_emojis:
  running: "runner"
  swimming: "swimmer"
  cycling: "bicyclist"
  strength_training: "muscle"
  yoga: "person_in_lotus_position"
  hiking: "hiking_boot"
```

Required fields: `garmin.email`, `garmin.password`, `slack.token`.  
All `rest_day` fields have defaults; `activity_emojis` is fully optional.

---

## Garmin Connect Client

Authentication uses the unofficial Garmin SSO flow (same as the browser):

1. `GET https://sso.garmin.com/sso/embed` — fetch CSRF token
2. `POST https://sso.garmin.com/sso/embed` with email + password → service ticket
3. `GET https://connect.garmin.com/modern` with ticket → session cookies
4. `GET https://connect.garmin.com/activitylist-service/activities/search/activities?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD` → activity list

No session is persisted between runs. Each crontab invocation performs a fresh login.

**Activity data used per entry:**
- `activityType.typeKey` — e.g. `running`, `strength_training`, `swimming`
- `duration` — seconds (converted to minutes for display)
- `distance` — metres (converted to km; omitted for non-distance activities)

---

## Formatter

**Smart truncation logic (Slack status text limit: 100 chars):**

1. Map each activity to emoji + human label + stats:
   - Distance activities: `🏊 Swimming 2.1km`
   - Duration-only activities: `🏋️ Strength 45min`
2. Join with ` & `: `🏋️ Strength 45min & 🏊 Swimming 2.1km`
3. If length > 100 → drop stats, keep labels: `🏋️ Strength & 🏊 Swimming`
4. If still > 100 → truncate last activity and append `…`

**Built-in emoji defaults (Garmin `typeKey` → Slack emoji name):**

| typeKey | emoji |
|---|---|
| running | runner |
| swimming | swimmer |
| cycling | bicyclist |
| strength_training | muscle |
| yoga | person_in_lotus_position |
| hiking | hiking_boot |
| walking | walking |
| elliptical | person_bouncing_ball |
| cardio | heart |
| rowing | rowing |
| skiing | skier |
| snowboarding | snowboarder |
| tennis | tennis |
| golf | golf |
| basketball | basketball |
| soccer | soccer |
| other | athletic_shoe |

The `status_emoji` field in the Slack API call (the main profile icon) is set to the emoji of the first activity of the day.

---

## Slack Client

```
POST https://slack.com/api/users.profile.set
Authorization: Bearer xoxp-...
Content-Type: application/json

{
  "profile": {
    "status_text": "🏋️ Strength & 🏊 Swimming",
    "status_emoji": ":muscle:",
    "status_expiration": 0
  }
}
```

`status_expiration: 0` — status does not auto-expire; the next crontab run overwrites it.

---

## Error Handling

| Situation | Behaviour |
|---|---|
| Config not found | Create template, print instructions, exit 0 |
| Invalid YAML or missing required field | Print field name, exit 1 |
| Garmin auth failure | `Error: Garmin login failed — check credentials`, exit 1 |
| No activities yesterday | Apply `rest_day` config |
| Slack API error | Print API response body, exit 1 |
| `--dry-run` | Skip Slack API call, print status to stdout |

All errors go to `stderr`. Normal runs are silent (crontab-friendly). `--dry-run` output goes to `stdout`.

---

## Versioning & CI/CD

### Commit convention (Conventional Commits)

| Prefix | Version bump |
|---|---|
| `fix:` | patch (1.0.0 → 1.0.1) |
| `feat:` | minor (1.0.0 → 1.1.0) |
| `feat!:` / `BREAKING CHANGE:` | major (1.0.0 → 2.0.0) |
| `chore:`, `docs:`, `refactor:` | no bump |

### Workflow

```
push to main
    ├─ ci.yml: go build + go test (matrix: linux, macOS, windows)
    └─ release-please: maintains Release PR with CHANGELOG.md

merge Release PR → tag vX.Y.Z created
    └─ release.yml: GoReleaser builds and publishes GitHub Release
```

### Build targets (GoReleaser)

| OS | Arch |
|---|---|
| linux | amd64, arm64 |
| darwin | amd64, arm64 |
| windows | amd64 |

Version is embedded at build time via `-ldflags "-X main.version={{.Version}}"` and exposed via `--version` flag.

---

## Crontab Setup (README)

```bash
# Install
go install github.com/plapko/garminslacknotify/cmd/garminslacknotify@latest
# or download binary from GitHub Releases and place in /usr/local/bin/

# First run — creates config template
garminslacknotify

# Edit config
nano ~/.config/garminslacknotify/config.yaml

# Test
garminslacknotify --dry-run

# Add to crontab (runs at 7:00 AM daily)
crontab -e
0 7 * * * /usr/local/bin/garminslacknotify >> /var/log/garminslacknotify.log 2>&1
```
