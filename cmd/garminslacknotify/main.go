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
