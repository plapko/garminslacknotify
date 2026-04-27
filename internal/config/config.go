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
	if c.RestDay.Text == "" && c.RestDay.Emoji == "" {
		// Whole rest_day section appears unconfigured — apply all defaults
		c.RestDay.Enabled = true
		c.RestDay.Text = "Rest day"
		c.RestDay.Emoji = "sleeping"
	} else if c.RestDay.Emoji == "" {
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
