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
