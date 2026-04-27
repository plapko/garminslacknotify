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
