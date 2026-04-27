package garmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"
)

const (
	defaultSSOBase     = "https://sso.garmin.com"
	defaultConnectBase = "https://connect.garmin.com"
)

// Activity holds the workout data needed for status formatting.
type Activity struct {
	TypeKey  string
	Duration float64 // seconds
	Distance float64 // metres, 0 if not applicable
}

// Client authenticates with Garmin Connect and fetches activities.
type Client struct {
	email       string
	password    string
	ssoBase     string
	connectBase string
	http        *http.Client
}

// New creates a Client using the real Garmin Connect endpoints.
func New(email, password string) *Client {
	return NewWithBaseURL(email, password, defaultSSOBase, defaultConnectBase)
}

// NewWithBaseURL creates a Client with custom base URLs (used in tests).
func NewWithBaseURL(email, password, ssoBase, connectBase string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		email:       email,
		password:    password,
		ssoBase:     ssoBase,
		connectBase: connectBase,
		http:        &http.Client{Jar: jar},
	}
}

// FetchActivities authenticates and returns activities for the given date.
func (c *Client) FetchActivities(date time.Time) ([]Activity, error) {
	if err := c.authenticate(); err != nil {
		return nil, err
	}
	return c.fetchActivities(date)
}

func (c *Client) authenticate() error {
	signinURL := c.ssoBase + "/sso/signin"
	params := url.Values{
		"service":   {c.connectBase + "/modern/"},
		"clientId":  {"GarminConnect"},
		"gauthHost": {c.ssoBase + "/sso"},
	}

	resp, err := c.http.Get(signinURL + "?" + params.Encode())
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}

	csrf := extractCSRF(body)

	form := url.Values{
		"username":  {c.email},
		"password":  {c.password},
		"embed":     {"true"},
		"_csrf":     {csrf},
		"service":   {c.connectBase + "/modern/"},
		"clientId":  {"GarminConnect"},
		"gauthHost": {c.ssoBase + "/sso"},
	}
	resp2, err := c.http.PostForm(signinURL, form)
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	body2, err := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}

	ticket := extractTicket(body2)
	if ticket == "" {
		return errors.New("garmin login failed — check credentials")
	}

	resp3, err := c.http.Get(c.connectBase + "/modern/?ticket=" + ticket)
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	resp3.Body.Close()
	return nil
}

func (c *Client) fetchActivities(date time.Time) ([]Activity, error) {
	dateStr := date.Format("2006-01-02")
	u := c.connectBase + "/activitylist-service/activities/search/activities"

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("startDate", dateStr)
	q.Set("endDate", dateStr)
	q.Set("limit", "100")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("NK", "NT")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw []struct {
		ActivityType struct {
			TypeKey string `json:"typeKey"`
		} `json:"activityType"`
		Duration float64 `json:"duration"`
		Distance float64 `json:"distance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	activities := make([]Activity, len(raw))
	for i, r := range raw {
		activities[i] = Activity{
			TypeKey:  r.ActivityType.TypeKey,
			Duration: r.Duration,
			Distance: r.Distance,
		}
	}
	return activities, nil
}

var csrfRe = regexp.MustCompile(`name="_csrf"\s+value="([^"]+)"`)
var ticketRe = regexp.MustCompile(`ticket=([A-Za-z0-9_\-]+)`)

func extractCSRF(body []byte) string {
	m := csrfRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}

func extractTicket(body []byte) string {
	m := ticketRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}
