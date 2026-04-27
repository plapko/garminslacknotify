package garmin_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/plapko/garminslacknotify/internal/garmin"
)

// mockGarminServer sets up a fake SSO + Connect server on a single httptest.Server.
// All endpoints (SSO signin, /modern/, activities) run on the same base URL.
func mockGarminServer(activities []map[string]interface{}) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/sso/signin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, `<input name="_csrf" value="test-csrf-token">`)
			return
		}
		// POST: redirect to /app/?ticket=... (modern flow, no embed=true)
		http.SetCookie(w, &http.Cookie{Name: "CASTGC", Value: "TGT-test"})
		http.Redirect(w, r, "/app/?ticket=ST-test-ticket", http.StatusFound)
	})

	mux.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "test-session"})
		// Serve a minimal app page with the CSRF meta tag.
		fmt.Fprint(w, `<html><head><meta name="csrf-token" content="test-app-csrf"/></head></html>`)
	})

	mux.HandleFunc("/gc-api/activitylist-service/activities/search/activities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(activities)
	})

	return httptest.NewServer(mux)
}

func TestFetchActivities_ReturnsActivities(t *testing.T) {
	raw := []map[string]interface{}{
		{
			"activityType":   map[string]interface{}{"typeKey": "running"},
			"duration":       1800.0,
			"distance":       5000.0,
			"startTimeLocal": "2026-04-26 10:00:00",
		},
		{
			"activityType":   map[string]interface{}{"typeKey": "strength_training"},
			"duration":       2700.0,
			"distance":       0.0,
			"startTimeLocal": "2026-04-26 12:00:00",
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

func TestFetchActivities_SessionCacheSkipsAuth(t *testing.T) {
	authCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/sso/signin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, `<input name="_csrf" value="csrf">`)
			return
		}
		authCalls++
		http.Redirect(w, r, "/app/?ticket=ST-test", http.StatusFound)
	})
	mux.HandleFunc("/app/", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "sess"})
		fmt.Fprint(w, `<html><head><meta name="csrf-token" content="tok"/></head></html>`)
	})
	raw := []map[string]interface{}{
		{"activityType": map[string]interface{}{"typeKey": "running"}, "duration": 600.0, "distance": 1000.0, "startTimeLocal": time.Now().Format("2006-01-02") + " 08:00:00"},
	}
	mux.HandleFunc("/gc-api/activitylist-service/activities/search/activities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(raw)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sessionFile := filepath.Join(t.TempDir(), "session.json")
	date := time.Now()

	// First call: authenticates and saves session.
	c1 := garmin.NewWithBaseURL("u@example.com", "p", srv.URL, srv.URL).SetSessionFile(sessionFile)
	if _, err := c1.FetchActivities(date); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if authCalls != 1 {
		t.Fatalf("expected 1 auth call after first fetch, got %d", authCalls)
	}

	// Second call: loads session from cache, no auth.
	c2 := garmin.NewWithBaseURL("u@example.com", "p", srv.URL, srv.URL).SetSessionFile(sessionFile)
	if _, err := c2.FetchActivities(date); err != nil {
		t.Fatalf("second call: %v", err)
	}
	if authCalls != 1 {
		t.Errorf("expected still 1 auth call after cached fetch, got %d", authCalls)
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
