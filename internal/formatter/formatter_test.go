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

func TestPrimaryEmoji_NormalizesSubtypes(t *testing.T) {
	cases := []struct {
		typeKey string
		want    string
	}{
		{"lap_swimming", "swimmer"},
		{"pool_swimming", "swimmer"},
		{"open_water_swimming", "swimmer"},
		{"treadmill_running", "runner"},
		{"trail_running", "runner"},
		{"track_running", "runner"},
		{"street_running", "runner"},
		{"indoor_running", "runner"},
		{"virtual_run", "runner"},
		{"ultra_run", "runner"},
		{"road_biking", "bicyclist"},
		{"mountain_biking", "bicyclist"},
		{"gravel_cycling", "bicyclist"},
		{"indoor_cycling", "bicyclist"},
		{"cyclocross", "bicyclist"},
		{"downhill_biking", "bicyclist"},
		{"virtual_ride", "bicyclist"},
		{"casual_walking", "walking"},
		{"speed_walking", "walking"},
		{"treadmill_walking", "walking"},
		{"indoor_rowing", "rowing"},
		{"resort_skiing", "skier"},
		{"backcountry_skiing", "skier"},
		{"cross_country_skiing", "skier"},
		{"resort_snowboarding", "snowboarder"},
		{"backcountry_snowboarding", "snowboarder"},
	}
	for _, tc := range cases {
		t.Run(tc.typeKey, func(t *testing.T) {
			activities := []garmin.Activity{{TypeKey: tc.typeKey}}
			got := formatter.PrimaryEmoji(activities, nil)
			if got != tc.want {
				t.Errorf("PrimaryEmoji(%q) = %q, want %q", tc.typeKey, got, tc.want)
			}
		})
	}
}

func TestPrimaryEmoji_CustomOverride_AppliesToSubtype(t *testing.T) {
	// User configures custom emoji for the parent type — it should also
	// apply to all subtypes (e.g. "swimming" override covers "lap_swimming").
	activities := []garmin.Activity{{TypeKey: "lap_swimming"}}
	result := formatter.PrimaryEmoji(activities, map[string]string{"swimming": "fish"})
	if result != "fish" {
		t.Errorf("got %q, want %q", result, "fish")
	}
}

func TestPrimaryEmoji_CustomOverride_ExactSubtypeBeatsParent(t *testing.T) {
	// If user explicitly overrides a subtype, that should win over the parent override.
	activities := []garmin.Activity{{TypeKey: "lap_swimming"}}
	result := formatter.PrimaryEmoji(activities, map[string]string{
		"lap_swimming": "tropical_fish",
		"swimming":     "fish",
	})
	if result != "tropical_fish" {
		t.Errorf("got %q, want %q", result, "tropical_fish")
	}
}

func TestPrimaryEmoji_UnknownSubtypeFallsBack(t *testing.T) {
	activities := []garmin.Activity{{TypeKey: "kitesurfing"}}
	result := formatter.PrimaryEmoji(activities, nil)
	if result != "athletic_shoe" {
		t.Errorf("got %q, want %q", result, "athletic_shoe")
	}
}

func TestFormatStatus_UsesParentEmojiForSubtype(t *testing.T) {
	activities := []garmin.Activity{
		{TypeKey: "lap_swimming", Duration: 1800, Distance: 2100},
	}
	result := formatter.FormatStatus(activities, nil)
	if !strings.Contains(result, "🏊") {
		t.Errorf("expected swim emoji in %q", result)
	}
}
