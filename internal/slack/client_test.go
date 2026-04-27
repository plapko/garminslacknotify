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
