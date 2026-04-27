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
// Tries to include stats; falls back to labels-only if > 100 runes.
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
