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

// mockGarminServer sets up a fake SSO + Connect server on a single httptest.Server.
// All endpoints (SSO signin, /modern/, activities) run on the same base URL.
func mockGarminServer(activities []map[string]interface{}) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/sso/signin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprint(w, `<input name="_csrf" value="test-csrf-token">`)
			return
		}
		// POST: set cookie and return page with ticket
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
