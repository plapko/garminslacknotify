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
	"github.com/pterm/pterm"
)

var version = "dev"

func main() {
	configPath := flag.String("config", defaultConfigPath(), "path to config file")
	dryRun := flag.Bool("dry-run", false, "print resulting status without setting it in Slack")
	dateStr := flag.String("date", "", "target date YYYY-MM-DD (default: yesterday)")
	showVersion := flag.Bool("version", false, "show version and exit")
	debug := flag.Bool("debug", false, "print HTTP requests/responses and auth details to stderr")
	flag.Usage = printUsage
	flag.Parse()

	if *showVersion {
		fmt.Printf("garminslacknotify %s\n", version)
		return
	}

	cfg, err := config.Load(*configPath)
	if os.IsNotExist(err) {
		if writeErr := config.WriteTemplate(*configPath); writeErr != nil {
			pterm.Error.Printf("could not create config: %v\n", writeErr)
			os.Exit(1)
		}
		pterm.Info.Printf("Config file created: %s\n", *configPath)
		pterm.Println("Edit it with your Garmin credentials and Slack token, then run again.")
		return
	}
	if err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}

	date := yesterday()
	if *dateStr != "" {
		parsed, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			pterm.Error.Println("invalid date format, use YYYY-MM-DD")
			os.Exit(1)
		}
		date = parsed
	}

	// Step 1: fetch Garmin activities
	spinner, _ := pterm.DefaultSpinner.Start("Connecting to Garmin Connect…")
	var garminClient *garmin.Client
	if *debug {
		garminClient = garmin.NewWithDebug(cfg.Garmin.Email, cfg.Garmin.Password, os.Stderr)
	} else {
		garminClient = garmin.New(cfg.Garmin.Email, cfg.Garmin.Password)
	}
	activities, err := garminClient.FetchActivities(date)
	if err != nil {
		spinner.Fail("Garmin Connect: " + err.Error())
		os.Exit(1)
	}
	n := len(activities)
	switch n {
	case 0:
		spinner.Success("Garmin Connect · no activities on " + date.Format("2006-01-02"))
	case 1:
		spinner.Success("Garmin Connect · 1 activity found")
	default:
		spinner.Success(fmt.Sprintf("Garmin Connect · %d activities found", n))
	}

	// Step 2: build status
	var statusText, statusEmoji string
	if n == 0 {
		if !cfg.RestDay.Enabled {
			pterm.Info.Println("Rest day status disabled — nothing to do.")
			return
		}
		statusText = cfg.RestDay.Text
		statusEmoji = cfg.RestDay.Emoji
	} else {
		statusText = formatter.FormatStatus(activities, cfg.ActivityEmojis)
		statusEmoji = formatter.PrimaryEmoji(activities, cfg.ActivityEmojis)
	}

	if *dryRun {
		pterm.Info.Println("Dry run — Slack status not changed.")
		pterm.Println()
		pterm.Println(pterm.Bold.Sprint("  Status text:  ") + statusText)
		pterm.Println(pterm.Bold.Sprint("  Status emoji: ") + ":" + statusEmoji + ":")
		return
	}

	// Step 3: set Slack status
	spinner2, _ := pterm.DefaultSpinner.Start("Setting Slack status…")
	var slackClient *slack.Client
	if *debug {
		slackClient = slack.NewWithDebug(cfg.Slack.Token, os.Stderr)
	} else {
		slackClient = slack.New(cfg.Slack.Token)
	}
	if err := slackClient.SetStatus(statusText, statusEmoji); err != nil {
		spinner2.Fail("Slack: " + err.Error())
		os.Exit(1)
	}
	spinner2.Success("Slack status set · " + statusText + "  :" + statusEmoji + ":")
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
  --debug     print HTTP requests/responses and auth details to stderr
  --help      show this help

Examples:
  garminslacknotify                      run with default config
  garminslacknotify --dry-run            preview status, don't set it
  garminslacknotify --date 2026-04-20    check a specific past date
  garminslacknotify --config ~/my.yaml   use custom config path
  garminslacknotify --debug 2>debug.log  save debug output to file

`)
}
