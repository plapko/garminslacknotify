# garminslacknotify Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool that fetches yesterday's Garmin Connect workouts and sets the Slack profile status with emoji-rich summary.

**Architecture:** Layered Go project — `cmd/garminslacknotify/main.go` orchestrates four independent `internal/` packages: `config` (YAML load/validate/template), `garmin` (SSO auth + activities API), `formatter` (emoji map + smart truncation), `slack` (profile status API). No shared state between packages; main wires them together.

**Tech Stack:** Go 1.22, `gopkg.in/yaml.v3`, standard library (`net/http`, `encoding/json`, `net/http/cookiejar`, `regexp`, `flag`). Tests use `net/http/httptest`. CI: GitHub Actions + GoReleaser + release-please.

**Beads epic:** `garminslacknotify-yhq`

---

## File Map

| File | Responsibility |
|---|---|
| `go.mod` | Module declaration, single external dependency |
| `cmd/garminslacknotify/main.go` | Flag parsing, orchestration, first-run config creation |
| `internal/config/config.go` | `Config` struct, `Load()`, `Validate()`, `WriteTemplate()` |
| `internal/config/config_test.go` | Config unit tests |
| `internal/garmin/client.go` | `Activity` struct, `Client`, SSO auth, `FetchActivities()` |
| `internal/garmin/client_test.go` | Garmin client tests via `httptest` mock server |
| `internal/formatter/formatter.go` | Unicode emoji map, Slack emoji map, `FormatStatus()`, `PrimaryEmoji()` |
| `internal/formatter/formatter_test.go` | Formatter unit tests |
| `internal/slack/client.go` | `Client`, `SetStatus()` |
| `internal/slack/client_test.go` | Slack client tests via `httptest` mock server |
| `.github/workflows/ci.yml` | Build + test on push/PR (linux, macOS, windows matrix) |
| `.github/workflows/release-please.yml` | Auto release PR + CHANGELOG on merge to main |
| `.github/workflows/release.yml` | GoReleaser triggered on `v*` tag |
| `.goreleaser.yaml` | Cross-platform build config |
| `release-please-config.json` | release-please Go project config |
| `.release-please-manifest.json` | Initial version manifest |
| `config.example.yaml` | Annotated example for users |
| `README.md` | Install, first-run, crontab setup |

---

## Task 1: Project scaffold

**Beads:** `bd create "Project scaffold: go.mod and directory structure" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Create: `go.mod`
- Create: `cmd/garminslacknotify/.gitkeep`
- Create: `internal/config/.gitkeep`
- Create: `internal/garmin/.gitkeep`
- Create: `internal/formatter/.gitkeep`
- Create: `internal/slack/.gitkeep`
- Create: `.github/workflows/.gitkeep`

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p cmd/garminslacknotify internal/config internal/garmin internal/formatter internal/slack .github/workflows
```

- [ ] **Step 2: Initialize Go module**

```bash
go mod init github.com/plapko/garminslacknotify
```

Expected: `go.mod` created with `module github.com/plapko/garminslacknotify` and `go 1.22` (or current Go version).

- [ ] **Step 3: Add yaml dependency**

```bash
go get gopkg.in/yaml.v3
```

Expected: `go.mod` updated with `require gopkg.in/yaml.v3 v3.0.1` (or latest), `go.sum` created.

- [ ] **Step 4: Verify build works on empty project**

Create `cmd/garminslacknotify/main.go` with minimal content:

```go
package main

func main() {}
```

Run:
```bash
go build ./...
```

Expected: exits 0, no errors.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum cmd/ internal/ .github/
git commit -m "chore: initialize Go module and project structure"
```

---

## Task 2: Config package

**Beads:** `bd create "Config package: Load, Validate, WriteTemplate" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plapko/garminslacknotify/internal/config"
)

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`
garmin:
  email: user@example.com
  password: secret
slack:
  token: xoxp-123
`), 0600)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Garmin.Email != "user@example.com" {
		t.Errorf("got email %q, want %q", cfg.Garmin.Email, "user@example.com")
	}
	if cfg.Garmin.Password != "secret" {
		t.Errorf("got password %q, want %q", cfg.Garmin.Password, "secret")
	}
	if cfg.Slack.Token != "xoxp-123" {
		t.Errorf("got token %q, want %q", cfg.Slack.Token, "xoxp-123")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected IsNotExist error, got: %v", err)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(`not: valid: yaml:`), 0600)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestValidate_MissingEmail(t *testing.T) {
	cfg := &config.Config{}
	cfg.Garmin.Password = "secret"
	cfg.Slack.Token = "xoxp-123"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing email")
	}
	if !strings.Contains(err.Error(), "garmin.email") {
		t.Errorf("expected error to mention garmin.email, got: %v", err)
	}
}

func TestValidate_MissingPassword(t *testing.T) {
	cfg := &config.Config{}
	cfg.Garmin.Email = "user@example.com"
	cfg.Slack.Token = "xoxp-123"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing password")
	}
	if !strings.Contains(err.Error(), "garmin.password") {
		t.Errorf("expected error to mention garmin.password, got: %v", err)
	}
}

func TestValidate_MissingToken(t *testing.T) {
	cfg := &config.Config{}
	cfg.Garmin.Email = "user@example.com"
	cfg.Garmin.Password = "secret"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "slack.token") {
		t.Errorf("expected error to mention slack.token, got: %v", err)
	}
}

func TestValidate_AppliesRestDayDefaults(t *testing.T) {
	cfg := &config.Config{}
	cfg.Garmin.Email = "user@example.com"
	cfg.Garmin.Password = "secret"
	cfg.Slack.Token = "xoxp-123"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.RestDay.Enabled {
		t.Error("expected RestDay.Enabled to default to true")
	}
	if cfg.RestDay.Text != "Rest day" {
		t.Errorf("got rest day text %q, want %q", cfg.RestDay.Text, "Rest day")
	}
	if cfg.RestDay.Emoji != "sleeping" {
		t.Errorf("got rest day emoji %q, want %q", cfg.RestDay.Emoji, "sleeping")
	}
}

func TestValidate_PreservesExplicitRestDay(t *testing.T) {
	cfg := &config.Config{}
	cfg.Garmin.Email = "user@example.com"
	cfg.Garmin.Password = "secret"
	cfg.Slack.Token = "xoxp-123"
	cfg.RestDay.Enabled = false
	cfg.RestDay.Text = "No workout"
	cfg.RestDay.Emoji = "zzz"

	cfg.Validate()

	if cfg.RestDay.Enabled {
		t.Error("expected RestDay.Enabled to stay false")
	}
	if cfg.RestDay.Text != "No workout" {
		t.Errorf("got rest day text %q, want %q", cfg.RestDay.Text, "No workout")
	}
}

func TestWriteTemplate_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := config.WriteTemplate(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read created file: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty template file")
	}
	content := string(data)
	if !strings.Contains(content, "garmin:") {
		t.Error("expected template to contain 'garmin:' section")
	}
	if !strings.Contains(content, "slack:") {
		t.Error("expected template to contain 'slack:' section")
	}
	if !strings.Contains(content, "xoxp-") {
		t.Error("expected template to mention xoxp- token format")
	}
}

func TestWriteTemplate_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "config.yaml")

	if err := config.WriteTemplate(path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/config/...
```

Expected: FAIL — `config` package does not exist yet.

- [ ] **Step 3: Implement config.go**

Create `internal/config/config.go`:

```go
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Garmin struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"garmin"`
	Slack struct {
		Token string `yaml:"token"`
	} `yaml:"slack"`
	RestDay struct {
		Enabled bool   `yaml:"enabled"`
		Text    string `yaml:"text"`
		Emoji   string `yaml:"emoji"`
	} `yaml:"rest_day"`
	ActivityEmojis map[string]string `yaml:"activity_emojis"`

	restDayExplicit bool
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Garmin.Email == "" {
		return errors.New("garmin.email is required")
	}
	if c.Garmin.Password == "" {
		return errors.New("garmin.password is required")
	}
	if c.Slack.Token == "" {
		return errors.New("slack.token is required")
	}
	if c.RestDay.Text == "" {
		c.RestDay.Enabled = true
		c.RestDay.Text = "Rest day"
		c.RestDay.Emoji = "sleeping"
	}
	return nil
}

const configTemplate = `# garminslacknotify configuration
#
# GARMIN CREDENTIALS
# Use your Garmin Connect login credentials (https://connect.garmin.com)
garmin:
  email: ""
  password: ""

# SLACK TOKEN
# 1. Go to https://api.slack.com/apps and create a new app (or use an existing one)
# 2. Under "OAuth & Permissions", add the User Token Scope: users.profile:write
# 3. Install the app to your workspace
# 4. Copy the "User OAuth Token" (starts with xoxp-)
slack:
  token: ""

# REST DAY STATUS
# What to set when no workouts are found for the target date.
# Set enabled: false to do nothing on rest days.
rest_day:
  enabled: true
  text: "Rest day"
  emoji: "sleeping"

# ACTIVITY EMOJI OVERRIDES (optional)
# Override the Slack profile icon emoji for specific Garmin activity types.
# Use Slack emoji names without colons (e.g. "muscle", "runner").
# Remove this section to use built-in defaults.
# activity_emojis:
#   running: runner
#   swimming: swimmer
#   strength_training: muscle
#   cycling: bicyclist
#   yoga: person_in_lotus_position
#   hiking: hiking_boot
`

func WriteTemplate(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(configTemplate), 0600)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/config/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config package with YAML load, validate, and template"
```

---

## Task 3: Formatter package

**Beads:** `bd create "Formatter package: emoji map and smart truncation" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Create: `internal/formatter/formatter.go`
- Create: `internal/formatter/formatter_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/formatter/formatter_test.go`:

```go
package formatter_test

import (
	"strings"
	"testing"

	"github.com/plapko/garminslacknotify/internal/formatter"
	"github.com/plapko/garminslacknotify/internal/garmin"
)

func TestFormatStatus_Empty(t *testing.T) {
	result := formatter.FormatStatus(nil, nil)
	if result != "" {
		t.Errorf("got %q, want empty string", result)
	}
}

func TestFormatStatus_SingleActivity_DurationOnly(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "strength_training", Duration: 2700, Distance: 0},
	}
	result := formatter.FormatStatus(activities, nil)
	if !strings.Contains(result, "45min") {
		t.Errorf("expected '45min' in %q", result)
	}
	if !strings.Contains(result, "🏋") {
		t.Errorf("expected strength emoji in %q", result)
	}
}

func TestFormatStatus_SingleActivity_WithDistance(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "swimming", Duration: 1800, Distance: 2100},
	}
	result := formatter.FormatStatus(activities, nil)
	if !strings.Contains(result, "2.1km") {
		t.Errorf("expected '2.1km' in %q", result)
	}
	if !strings.Contains(result, "🏊") {
		t.Errorf("expected swim emoji in %q", result)
	}
}

func TestFormatStatus_TwoActivities(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "strength_training", Duration: 2700, Distance: 0},
		{TypeKey: "swimming", Duration: 1800, Distance: 2100},
	}
	result := formatter.FormatStatus(activities, nil)
	if !strings.Contains(result, " & ") {
		t.Errorf("expected ' & ' separator in %q", result)
	}
	if len([]rune(result)) > 100 {
		t.Errorf("result exceeds 100 chars: %d chars in %q", len([]rune(result)), result)
	}
}

func TestFormatStatus_FallsBackWhenTooLong(t *testing.T) {
	// 6 activities with long stats will exceed 100 chars
	activities := []garmin.Activity{
		{TypeKey: "strength_training", Duration: 3600, Distance: 0},
		{TypeKey: "swimming", Duration: 3600, Distance: 10000},
		{TypeKey: "running", Duration: 3600, Distance: 10000},
		{TypeKey: "cycling", Duration: 3600, Distance: 100000},
		{TypeKey: "yoga", Duration: 3600, Distance: 0},
		{TypeKey: "hiking", Duration: 3600, Distance: 5000},
	}
	result := formatter.FormatStatus(activities, nil)
	if len([]rune(result)) > 100 {
		t.Errorf("result exceeds 100 chars: %d chars in %q", len([]rune(result)), result)
	}
}

func TestFormatStatus_CustomEmoji(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "running", Duration: 1800, Distance: 5000},
	}
	result := formatter.FormatStatus(activities, map[string]string{"running": "dash"})
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestFormatStatus_UnknownTypeKey(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "kitesurfing", Duration: 3600, Distance: 0},
	}
	result := formatter.FormatStatus(activities, nil)
	if result == "" {
		t.Error("expected non-empty result for unknown activity type")
	}
}

func TestPrimaryEmoji_ReturnsFirstActivity(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "strength_training"},
		{TypeKey: "swimming"},
	}
	result := formatter.PrimaryEmoji(activities, nil)
	if result != "muscle" {
		t.Errorf("got %q, want %q", result, "muscle")
	}
}

func TestPrimaryEmoji_CustomOverride(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "running"},
	}
	result := formatter.PrimaryEmoji(activities, map[string]string{"running": "dash"})
	if result != "dash" {
		t.Errorf("got %q, want %q", result, "dash")
	}
}

func TestPrimaryEmoji_Empty(t *testing.T) {
	result := formatter.PrimaryEmoji(nil, nil)
	if result != "" {
		t.Errorf("got %q, want empty string", result)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/formatter/...
```

Expected: FAIL — `formatter` package does not exist yet.

- [ ] **Step 3: Implement formatter.go**

Create `internal/formatter/formatter.go`:

```go
package formatter

import (
	"fmt"
	"strings"

	"github.com/plapko/garminslacknotify/internal/garmin"
)

const statusLimit = 100

var unicodeEmojis = map[string]string{
	"running":           "🏃",
	"swimming":          "🏊",
	"strength_training": "🏋️",
	"cycling":           "🚴",
	"yoga":              "🧘",
	"hiking":            "🥾",
	"walking":           "🚶",
	"elliptical":        "🏃",
	"cardio":            "❤️",
	"rowing":            "🚣",
	"skiing":            "⛷️",
	"snowboarding":      "🏂",
	"tennis":            "🎾",
	"golf":              "⛳",
	"basketball":        "🏀",
	"soccer":            "⚽",
}

var slackEmojis = map[string]string{
	"running":           "runner",
	"swimming":          "swimmer",
	"strength_training": "muscle",
	"cycling":           "bicyclist",
	"yoga":              "person_in_lotus_position",
	"hiking":            "hiking_boot",
	"walking":           "walking",
	"elliptical":        "person_bouncing_ball",
	"cardio":            "heart",
	"rowing":            "rowing",
	"skiing":            "skier",
	"snowboarding":      "snowboarder",
	"tennis":            "tennis",
	"golf":              "golf",
	"basketball":        "basketball",
	"soccer":            "soccer",
}

var humanLabels = map[string]string{
	"running":           "Running",
	"swimming":          "Swimming",
	"strength_training": "Strength",
	"cycling":           "Cycling",
	"yoga":              "Yoga",
	"hiking":            "Hiking",
	"walking":           "Walking",
	"elliptical":        "Elliptical",
	"cardio":            "Cardio",
	"rowing":            "Rowing",
	"skiing":            "Skiing",
	"snowboarding":      "Snowboarding",
	"tennis":            "Tennis",
	"golf":              "Golf",
	"basketball":        "Basketball",
	"soccer":            "Soccer",
}

// FormatStatus builds the Slack status text from activities.
// It tries to include stats; falls back to labels-only if > 100 runes.
func FormatStatus(activities []garmin.Activity, customEmojis map[string]string) string {
	if len(activities) == 0 {
		return ""
	}

	full := joinParts(activities, true)
	if runeLen(full) <= statusLimit {
		return full
	}

	short := joinParts(activities, false)
	if runeLen(short) <= statusLimit {
		return short
	}

	return truncate(short, statusLimit)
}

// PrimaryEmoji returns the Slack emoji name for the first activity.
// Used as the profile status_emoji field.
func PrimaryEmoji(activities []garmin.Activity, customEmojis map[string]string) string {
	if len(activities) == 0 {
		return ""
	}
	typeKey := activities[0].TypeKey
	if customEmojis != nil {
		if e, ok := customEmojis[typeKey]; ok {
			return e
		}
	}
	if e, ok := slackEmojis[typeKey]; ok {
		return e
	}
	return "athletic_shoe"
}

func joinParts(activities []garmin.Activity, withStats bool) string {
	parts := make([]string, 0, len(activities))
	for _, a := range activities {
		emoji := unicodeEmoji(a.TypeKey)
		label := humanLabel(a.TypeKey)
		if withStats {
			stats := formatStats(a)
			if stats != "" {
				parts = append(parts, emoji+" "+label+" "+stats)
				continue
			}
		}
		parts = append(parts, emoji+" "+label)
	}
	return strings.Join(parts, " & ")
}

func formatStats(a garmin.Activity) string {
	if a.Distance > 0 {
		return fmt.Sprintf("%.1fkm", a.Distance/1000)
	}
	if a.Duration > 0 {
		return fmt.Sprintf("%dmin", int(a.Duration/60))
	}
	return ""
}

func unicodeEmoji(typeKey string) string {
	if e, ok := unicodeEmojis[typeKey]; ok {
		return e
	}
	return "🏅"
}

func humanLabel(typeKey string) string {
	if l, ok := humanLabels[typeKey]; ok {
		return l
	}
	words := strings.Split(typeKey, "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func runeLen(s string) int {
	return len([]rune(s))
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/formatter/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/formatter/
git commit -m "feat: add formatter package with emoji map and smart truncation"
```

---

## Task 4: Slack client package

**Beads:** `bd create "Slack client: users.profile.set" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Create: `internal/slack/client.go`
- Create: `internal/slack/client_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/slack/client_test.go`:

```go
package slack_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/plapko/garminslacknotify/internal/slack"
)

func TestSetStatus_SendsCorrectPayload(t *testing.T) {
	var gotBody map[string]interface{}
	var gotAuth string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer srv.Close()

	client := slack.NewWithBaseURL("xoxp-test-token", srv.URL)
	if err := client.SetStatus("🏋️ Strength 45min", "muscle"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth != "Bearer xoxp-test-token" {
		t.Errorf("got Authorization %q, want %q", gotAuth, "Bearer xoxp-test-token")
	}

	profile, ok := gotBody["profile"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected profile key in body, got: %v", gotBody)
	}
	if profile["status_text"] != "🏋️ Strength 45min" {
		t.Errorf("got status_text %v, want %q", profile["status_text"], "🏋️ Strength 45min")
	}
	if profile["status_emoji"] != ":muscle:" {
		t.Errorf("got status_emoji %v, want %q", profile["status_emoji"], ":muscle:")
	}
	if profile["status_expiration"] != float64(0) {
		t.Errorf("got status_expiration %v, want 0", profile["status_expiration"])
	}
}

func TestSetStatus_ReturnsErrorOnAPIFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "error": "invalid_auth"})
	}))
	defer srv.Close()

	client := slack.NewWithBaseURL("bad-token", srv.URL)
	err := client.SetStatus("test", "muscle")
	if err == nil {
		t.Fatal("expected error from API failure")
	}
	if !strings.Contains(err.Error(), "invalid_auth") {
		t.Errorf("expected error to mention 'invalid_auth', got: %v", err)
	}
}

func TestSetStatus_ReturnsErrorOnHTTPFailure(t *testing.T) {
	client := slack.NewWithBaseURL("token", "http://127.0.0.1:1")
	err := client.SetStatus("test", "muscle")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/slack/...
```

Expected: FAIL — `slack` package does not exist yet.

- [ ] **Step 3: Implement slack/client.go**

Create `internal/slack/client.go`:

```go
package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://slack.com/api"

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

func New(token string) *Client {
	return NewWithBaseURL(token, defaultBaseURL)
}

func NewWithBaseURL(token, baseURL string) *Client {
	return &Client{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *Client) SetStatus(text, emoji string) error {
	payload := map[string]interface{}{
		"profile": map[string]interface{}{
			"status_text":       text,
			"status_emoji":      ":" + emoji + ":",
			"status_expiration": 0,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/users.profile.set", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("slack API error: %s", result.Error)
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/slack/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/slack/
git commit -m "feat: add Slack client for profile status"
```

---

## Task 5: Garmin client package

**Beads:** `bd create "Garmin client: SSO auth and activities fetch" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Create: `internal/garmin/client.go`
- Create: `internal/garmin/client_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/garmin/client_test.go`:

```go
package garmin_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/plapko/garminslacknotify/internal/garmin"
)

// mockGarminServer sets up a fake SSO + Connect server.
// It expects: GET /sso/signin, POST /sso/signin, GET /modern/, GET /activities
func mockGarminServer(activities []map[string]interface{}) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/sso/signin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, `<input name="_csrf" value="test-csrf-token">`)
			return
		}
		// POST: set cookie and return ticket
		http.SetCookie(w, &http.Cookie{Name: "CASTGC", Value: "TGT-test"})
		fmt.Fprint(w, `<a href="/modern/?ticket=ST-test-ticket">continue</a>`)
	})

	mux.HandleFunc("/modern/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "SESSIONID", Value: "test-session"})
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/activitylist-service/activities/search/activities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(activities)
	})

	return httptest.NewServer(mux)
}

func TestFetchActivities_ReturnsActivities(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"activityType": map[string]interface{}{"typeKey": "running"},
			"duration":     1800.0,
			"distance":     5000.0,
		},
		{
			"activityType": map[string]interface{}{"typeKey": "strength_training"},
			"duration":     2700.0,
			"distance":     0.0,
		},
	}
	srv := mockGarminServer(raw)
	defer srv.Close()

	client := garmin.NewWithBaseURL("user@example.com", "password", srv.URL, srv.URL)
	date := time.Date(2026, 4, 26, 0, 0, 0, 0, time.UTC)

	activities, err := client.FetchActivities(date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("got %d activities, want 2", len(activities))
	}
	if activities[0].TypeKey != "running" {
		t.Errorf("got typeKey %q, want %q", activities[0].TypeKey, "running")
	}
	if activities[0].Duration != 1800 {
		t.Errorf("got duration %v, want 1800", activities[0].Duration)
	}
	if activities[0].Distance != 5000 {
		t.Errorf("got distance %v, want 5000", activities[0].Distance)
	}
	if activities[1].TypeKey != "strength_training" {
		t.Errorf("got typeKey %q, want %q", activities[1].TypeKey, "strength_training")
	}
}

func TestFetchActivities_ReturnsEmptySlice(t *testing.T) {
	srv := mockGarminServer([]map[string]interface{}{})
	defer srv.Close()

	client := garmin.NewWithBaseURL("user@example.com", "password", srv.URL, srv.URL)
	activities, err := client.FetchActivities(time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(activities) != 0 {
		t.Errorf("got %d activities, want 0", len(activities))
	}
}

func TestFetchActivities_AuthFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/sso/signin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, `<input name="_csrf" value="csrf">`)
			return
		}
		// No ticket in response — simulates bad credentials
		fmt.Fprint(w, `<p>Invalid credentials</p>`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := garmin.NewWithBaseURL("bad@example.com", "wrongpass", srv.URL, srv.URL)
	_, err := client.FetchActivities(time.Now())
	if err == nil {
		t.Fatal("expected error for auth failure")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/garmin/...
```

Expected: FAIL — `garmin` package does not exist yet.

- [ ] **Step 3: Implement garmin/client.go**

Create `internal/garmin/client.go`:

```go
package garmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"
)

const (
	defaultSSOBase     = "https://sso.garmin.com"
	defaultConnectBase = "https://connect.garmin.com"
)

// Activity holds the workout data needed for status formatting.
type Activity struct {
	TypeKey  string
	Duration float64 // seconds
	Distance float64 // metres, 0 if not applicable
}

// Client authenticates with Garmin Connect and fetches activities.
type Client struct {
	email       string
	password    string
	ssoBase     string
	connectBase string
	http        *http.Client
}

// New creates a Client using the real Garmin Connect endpoints.
func New(email, password string) *Client {
	return NewWithBaseURL(email, password, defaultSSOBase, defaultConnectBase)
}

// NewWithBaseURL creates a Client with custom base URLs (used in tests).
func NewWithBaseURL(email, password, ssoBase, connectBase string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		email:       email,
		password:    password,
		ssoBase:     ssoBase,
		connectBase: connectBase,
		http:        &http.Client{Jar: jar},
	}
}

// FetchActivities authenticates and returns activities for the given date.
func (c *Client) FetchActivities(date time.Time) ([]Activity, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}
	return c.fetchActivities(date)
}

func (c *Client) authenticate() error {
	signinURL := c.ssoBase + "/sso/signin"
	params := url.Values{
		"service":   {c.connectBase + "/modern/"},
		"clientId":  {"GarminConnect"},
		"gauthHost": {c.ssoBase + "/sso"},
	}

	resp, err := c.http.Get(signinURL + "?" + params.Encode())
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}

	csrf := extractCSRF(body)

	form := url.Values{
		"username":  {c.email},
		"password":  {c.password},
		"embed":     {"true"},
		"_csrf":     {csrf},
		"service":   {c.connectBase + "/modern/"},
		"clientId":  {"GarminConnect"},
		"gauthHost": {c.ssoBase + "/sso"},
	}
	resp2, err := c.http.PostForm(signinURL, form)
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	body2, err := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}

	ticket := extractTicket(body2)
	if ticket == "" {
		return errors.New("garmin login failed — check credentials")
	}

	resp3, err := c.http.Get(c.connectBase + "/modern/?ticket=" + ticket)
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	resp3.Body.Close()
	return nil
}

func (c *Client) fetchActivities(date time.Time) ([]Activity, error) {
	dateStr := date.Format("2006-01-02")
	u := c.connectBase + "/activitylist-service/activities/search/activities"

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("startDate", dateStr)
	q.Set("endDate", dateStr)
	q.Set("limit", "100")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("NK", "NT")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw []struct {
		ActivityType struct {
			TypeKey string `json:"typeKey"`
		} `json:"activityType"`
		Duration float64 `json:"duration"`
		Distance float64 `json:"distance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	activities := make([]Activity, len(raw))
	for i, r := range raw {
		activities[i] = Activity{
			TypeKey:  r.ActivityType.TypeKey,
			Duration: r.Duration,
			Distance: r.Distance,
		}
	}
	return activities, nil
}

var csrfRe = regexp.MustCompile(`name="_csrf"\s+value="([^"]+)"`)
var ticketRe = regexp.MustCompile(`ticket=([A-Za-z0-9_\-]+)`)

func extractCSRF(body []byte) string {
	m := csrfRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}

func extractTicket(body []byte) string {
	m := ticketRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/garmin/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/garmin/
git commit -m "feat: add Garmin Connect client with SSO auth and activity fetch"
```

---

## Task 6: Main CLI

**Beads:** `bd create "Main CLI: orchestration and flag parsing" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Modify: `cmd/garminslacknotify/main.go`

- [ ] **Step 1: Write the full main.go**

Replace `cmd/garminslacknotify/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/plapko/garminslacknotify/internal/config"
	"github.com/plapko/garminslacknotify/internal/formatter"
	"github.com/plapko/garminslacknotify/internal/garmin"
	"github.com/plapko/garminslacknotify/internal/slack"
)

var version = "dev"

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	dryRun := flag.Bool("dry-run", false, "print resulting status without setting it in Slack")
	dateStr := flag.String("date", "", "target date YYYY-MM-DD (default: yesterday)")
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Usage = printUsage
	flag.Parse()

	if *showVersion {
		fmt.Printf("garminslacknotify %s\n", version)
		return
	}

	cfg, err := config.Load(*configPath)
	if os.IsNotExist(err) {
		if writeErr := config.WriteTemplate(*configPath); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create config: %v\n", writeErr)
			os.Exit(1)
		}
		fmt.Printf("Config file created: %s\n", *configPath)
		fmt.Println("Edit it with your Garmin credentials and Slack token, then run again.")
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	date := yesterday()
	if *dateStr != "" {
		parsed, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date format, use YYYY-MM-DD\n")
			os.Exit(1)
		}
		date = parsed
	}

	garminClient := garmin.New(cfg.Garmin.Email, cfg.Garmin.Password)
	activities, err := garminClient.FetchActivities(date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var statusText, statusEmoji string
	if len(activities) == 0 {
		if !cfg.RestDay.Enabled {
			return
		}
		statusText = cfg.RestDay.Text
		statusEmoji = cfg.RestDay.Emoji
	} else {
		statusText = formatter.FormatStatus(activities, cfg.ActivityEmojis)
		statusEmoji = formatter.PrimaryEmoji(activities, cfg.ActivityEmojis)
	}

	if *dryRun {
		fmt.Printf("Status text:  %s\n", statusText)
		fmt.Printf("Status emoji: :%s:\n", statusEmoji)
		return
	}

	slackClient := slack.New(cfg.Slack.Token)
	if err := slackClient.SetStatus(statusText, statusEmoji); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func yesterday() time.Time {
	y, m, d := time.Now().Date()
	return time.Date(y, m, d-1, 0, 0, 0, 0, time.Local)
}

func defaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "garminslacknotify", "config.yaml")
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage: garminslacknotify [flags]

Flags:
  --config    path to config file
              (default: ~/.config/garminslacknotify/config.yaml)
  --dry-run   print resulting status without setting it in Slack
  --date      override target date (YYYY-MM-DD, default: yesterday)
  --version   show version and exit
  --help      show this help

Examples:
  garminslacknotify                      run with default config
  garminslacknotify --dry-run            preview status, don't set it
  garminslacknotify --date 2026-04-20    check a specific past date
  garminslacknotify --config ~/my.yaml   use custom config path

`)
}
```

- [ ] **Step 2: Build and smoke-test**

```bash
go build ./...
```

Expected: exits 0, binary created at `./garminslacknotify` (or `garminslacknotify.exe` on Windows).

```bash
./garminslacknotify --help
```

Expected: usage text printed to stderr, exits 0.

```bash
./garminslacknotify --version
```

Expected: `garminslacknotify dev`

- [ ] **Step 3: Test first-run config creation**

```bash
./garminslacknotify --config /tmp/test-gsn-config.yaml
```

Expected:
```
Config file created: /tmp/test-gsn-config.yaml
Edit it with your Garmin credentials and Slack token, then run again.
```

Verify file exists and contains the template:
```bash
cat /tmp/test-gsn-config.yaml
```

Expected: YAML with `garmin:`, `slack:`, `rest_day:` sections and inline comments.

- [ ] **Step 4: Run all tests**

```bash
go test ./...
```

Expected: all packages PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/garminslacknotify/main.go
git commit -m "feat: add CLI entry point with flag parsing and first-run config"
```

---

## Task 7: CI/CD — GitHub Actions, GoReleaser, release-please

**Beads:** `bd create "CI/CD: GitHub Actions, GoReleaser, release-please" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release-please.yml`
- Create: `.github/workflows/release.yml`
- Create: `.goreleaser.yaml`
- Create: `release-please-config.json`
- Create: `.release-please-manifest.json`

- [ ] **Step 1: Create CI workflow**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    name: Build and test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Build
        run: go build ./...

      - name: Test
        run: go test ./...
```

- [ ] **Step 2: Create release-please workflow**

Create `.github/workflows/release-please.yml`:

```yaml
name: Release Please

on:
  push:
    branches: [main]

permissions:
  contents: write
  pull-requests: write

jobs:
  release-please:
    runs-on: ubuntu-latest
    steps:
      - uses: googleapis/release-please-action@v4
        with:
          config-file: release-please-config.json
          manifest-file: .release-please-manifest.json
```

- [ ] **Step 3: Create GoReleaser release workflow**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 4: Create GoReleaser config**

Create `.goreleaser.yaml`:

```yaml
version: 2

project_name: garminslacknotify

before:
  hooks:
    - go mod tidy

builds:
  - id: garminslacknotify
    main: ./cmd/garminslacknotify
    binary: garminslacknotify
    ldflags:
      - -s -w -X main.version={{.Version}}
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ignore:
      - goos: windows
        goarch: arm64

archives:
  - id: default
    format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - config.example.yaml

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  use: github
```

- [ ] **Step 5: Create release-please config**

Create `release-please-config.json`:

```json
{
  "$schema": "https://raw.githubusercontent.com/googleapis/release-please/main/schemas/config.json",
  "release-type": "go",
  "packages": {
    ".": {}
  },
  "changelog-sections": [
    {"type": "feat", "section": "Features"},
    {"type": "fix", "section": "Bug Fixes"},
    {"type": "refactor", "section": "Refactoring"},
    {"type": "chore", "section": "Miscellaneous", "hidden": true},
    {"type": "docs", "section": "Documentation", "hidden": true}
  ]
}
```

Create `.release-please-manifest.json`:

```json
{
  ".": "0.1.0"
}
```

- [ ] **Step 6: Commit**

```bash
git add .github/ .goreleaser.yaml release-please-config.json .release-please-manifest.json
git commit -m "ci: add GitHub Actions workflows, GoReleaser, and release-please"
```

---

## Task 8: Documentation and example config

**Beads:** `bd create "Docs: README and config.example.yaml" -t task && bd dep add <id> garminslacknotify-yhq --type parent-child`

**Files:**
- Modify: `README.md`
- Create: `config.example.yaml`

- [ ] **Step 1: Create annotated example config**

Create `config.example.yaml`:

```yaml
# garminslacknotify example configuration
# Copy to ~/.config/garminslacknotify/config.yaml and fill in your credentials.
# Or run `garminslacknotify` once — it creates this file automatically.

# GARMIN CREDENTIALS
# Your Garmin Connect login (https://connect.garmin.com)
garmin:
  email: "user@example.com"
  password: "your-garmin-password"

# SLACK TOKEN
# 1. Go to https://api.slack.com/apps → Create New App → From scratch
# 2. Name it anything (e.g. "garminslacknotify"), pick your workspace
# 3. Go to "OAuth & Permissions" → under "User Token Scopes" add: users.profile:write
# 4. Click "Install to Workspace" → copy the "User OAuth Token" (starts with xoxp-)
slack:
  token: "xoxp-your-token-here"

# REST DAY STATUS
# Shown when no workouts found for the target date.
# Set enabled: false to do nothing on rest days.
rest_day:
  enabled: true
  text: "Rest day"
  emoji: "sleeping"

# ACTIVITY EMOJI OVERRIDES (optional)
# Override the Slack profile icon emoji per Garmin activity type.
# Use Slack emoji names without colons.
# Full list of Slack emoji: https://emojipedia.org/slack
# Remove or comment out this section to use built-in defaults.
activity_emojis:
  running: runner
  swimming: swimmer
  strength_training: muscle
  cycling: bicyclist
  yoga: person_in_lotus_position
  hiking: hiking_boot
```

- [ ] **Step 2: Write README.md**

Replace `README.md`:

```markdown
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
0 7 * * * /usr/local/bin/garminslacknotify >> /var/log/garminslacknotify.log 2>&1
```

## Usage

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
  garminslacknotify                      run with default config
  garminslacknotify --dry-run            preview status, don't set it
  garminslacknotify --date 2026-04-20    check a specific past date
  garminslacknotify --config ~/my.yaml   use custom config path
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
```

- [ ] **Step 3: Final build and test**

```bash
go build ./...
go test ./...
```

Expected: exits 0, all tests PASS.

- [ ] **Step 4: Commit**

```bash
git add README.md config.example.yaml
git commit -m "docs: add README and annotated example config"
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Covered by |
|---|---|
| Fetch yesterday's Garmin workouts | Task 5: `garmin.FetchActivities(date)` |
| Slack profile status (not channel) | Task 4: `users.profile.set` |
| Smart truncation (stats → labels → truncate) | Task 3: `formatter.FormatStatus` |
| YAML config with defaults | Task 2: `config.Validate` applies defaults |
| First-run auto config creation | Task 6: `os.IsNotExist` branch in main |
| --help with examples | Task 6: `printUsage()` |
| --dry-run | Task 6: skips Slack call, prints to stdout |
| --date override | Task 6: parsed in main |
| --version | Task 6: `showVersion` flag |
| rest_day configurable, default 😴 | Task 2: defaults in `Validate`, Task 6: orchestration |
| Errors to stderr, silent on success | Task 6: all `fmt.Fprintf(os.Stderr, ...)` |
| GitHub Actions CI | Task 7: ci.yml |
| release-please | Task 7: release-please.yml + config |
| GoReleaser 5 platforms | Task 7: .goreleaser.yaml |
| Version embedded via ldflags | Task 7: `.goreleaser.yaml` ldflags |
| README with crontab setup | Task 8 |
| config.example.yaml | Task 8 |
