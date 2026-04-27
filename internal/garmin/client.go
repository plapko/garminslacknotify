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
	"strings"
	"time"

	"github.com/plapko/garminslacknotify/internal/httpdebug"
)

const (
	defaultSSOBase     = "https://sso.garmin.com"
	defaultConnectBase = "https://connect.garmin.com"
	userAgent          = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
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
	debug       io.Writer
	appCSRF     string // CSRF token from the Garmin Connect app page
	jwtFGP      string // JWT fingerprint cookie (path-scoped to /app/, must be added manually to /proxy/ requests)
}

// New creates a Client using the real Garmin Connect endpoints.
func New(email, password string) *Client {
	return build(email, password, defaultSSOBase, defaultConnectBase, nil)
}

// NewWithBaseURL creates a Client with custom base URLs (used in tests).
func NewWithBaseURL(email, password, ssoBase, connectBase string) *Client {
	return build(email, password, ssoBase, connectBase, nil)
}

// NewWithDebug creates a Client that writes HTTP and auth debug info to debug.
func NewWithDebug(email, password string, debug io.Writer) *Client {
	return build(email, password, defaultSSOBase, defaultConnectBase, debug)
}

func build(email, password, ssoBase, connectBase string, debug io.Writer) *Client {
	jar, _ := cookiejar.New(nil)
	var base http.RoundTripper = http.DefaultTransport
	if debug != nil {
		base = &httpdebug.Transport{Base: base, Out: debug, Label: "garmin"}
	}
	return &Client{
		email:       email,
		password:    password,
		ssoBase:     ssoBase,
		connectBase: connectBase,
		http: &http.Client{
			Jar:       jar,
			Transport: &browserTransport{base: base},
		},
		debug: debug,
	}
}

// browserTransport adds a browser User-Agent to every request so Garmin
// does not reject the connection at the network layer.
type browserTransport struct {
	base http.RoundTripper
}

func (t *browserTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("User-Agent", userAgent)
	return t.base.RoundTrip(r)
}

func (c *Client) debugf(format string, args ...any) {
	if c.debug != nil {
		fmt.Fprintf(c.debug, "[debug] garmin   "+format+"\n", args...)
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
	if csrf == "" {
		c.debugf("CSRF token: not found in signin page")
	} else {
		c.debugf("CSRF token: found")
	}

	form := url.Values{
		"username":  {c.email},
		"password":  {c.password},
		"embed":     {"true"},
		"_csrf":     {csrf},
		"service":   {c.connectBase + "/modern/"},
		"clientId":  {"GarminConnect"},
		"gauthHost": {c.ssoBase + "/sso"},
	}
	req2, err := http.NewRequest(http.MethodPost, signinURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Origin", c.ssoBase)
	req2.Header.Set("Referer", signinURL+"?"+params.Encode())

	// Modern Garmin SSO redirects POST → 302 → connect.garmin.com/modern/?ticket=ST-XXX.
	// Go's http.Client follows the redirect automatically (setting session cookies),
	// so we capture the ticket URL via CheckRedirect before the body is consumed.
	var ticketFoundInRedirect bool
	c.http.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
		if extractTicket([]byte(req.URL.String())) != "" {
			ticketFoundInRedirect = true
			c.debugf("ticket: found in redirect → %s", req.URL)
		}
		return nil
	}
	resp2, err := c.http.Do(req2)
	c.http.CheckRedirect = nil
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	body2, err := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}

	if resp2.StatusCode == http.StatusTooManyRequests {
		c.debugf("login POST returned 429 — rate limited")
		return errors.New("garmin login failed: rate limited (429) — wait a few minutes and try again")
	}

	// Modern flow: ticket was in redirect URL; Go already followed it and set cookies.
	// body2 is the final /app/ page — extract its CSRF token for API requests.
	if ticketFoundInRedirect {
		c.appCSRF = extractAppCSRF(body2)
		c.captureJWTFGP()
		c.debugf("auth complete via redirect flow (app CSRF: %v, JWT_FGP: %v)", c.appCSRF != "", c.jwtFGP != "")
		return nil
	}

	// Legacy embed=true flow: ticket embedded in response body.
	ticket := extractTicket(body2)
	if ticket == "" {
		c.debugf("ticket: not found (neither in redirect nor body) — status %d, wrong credentials?", resp2.StatusCode)
		return errors.New("garmin login failed — check credentials")
	}
	c.debugf("ticket: found in response body (legacy flow)")

	resp3, err := c.http.Get(c.connectBase + "/modern/?ticket=" + ticket)
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	appBody, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	c.appCSRF = extractAppCSRF(appBody)
	c.captureJWTFGP()
	c.debugf("auth complete via body/ticket flow (app CSRF: %v, JWT_FGP: %v)", c.appCSRF != "", c.jwtFGP != "")
	return nil
}

// captureJWTFGP reads JWT_FGP from the cookie jar scoped to /app/.
// The cookie is path-restricted and won't be sent automatically to /proxy/ endpoints.
func (c *Client) captureJWTFGP() {
	appURL, _ := url.Parse(c.connectBase + "/app/")
	for _, cookie := range c.http.Jar.Cookies(appURL) {
		if cookie.Name == "JWT_FGP" {
			c.jwtFGP = cookie.Value
			return
		}
	}
}

func (c *Client) fetchActivities(date time.Time) ([]Activity, error) {
	dateStr := date.Format("2006-01-02")
	u := c.connectBase + "/gc-api/activitylist-service/activities/search/activities"

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("limit", "100")
	q.Set("start", "0")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if c.appCSRF != "" {
		req.Header.Set("connect-csrf-token", c.appCSRF)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Garmin returns {} (empty object) or "null" when there are no activities.
	if s := strings.TrimSpace(string(body)); s == "{}" || s == "null" {
		c.debugf("activities: empty response (%s)", s)
		return []Activity{}, nil
	}

	var raw []struct {
		ActivityType struct {
			TypeKey string `json:"typeKey"`
		} `json:"activityType"`
		Duration       float64 `json:"duration"`
		Distance       float64 `json:"distance"`
		StartTimeLocal string  `json:"startTimeLocal"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return nil, fmt.Errorf("garmin activities: HTTP %d, unexpected response: %s", resp.StatusCode, snippet)
	}

	c.debugf("activities: fetched %d items, filtering for %s", len(raw), dateStr)
	activities := make([]Activity, 0, len(raw))
	for _, r := range raw {
		if !strings.HasPrefix(r.StartTimeLocal, dateStr) {
			continue
		}
		activities = append(activities, Activity{
			TypeKey:  r.ActivityType.TypeKey,
			Duration: r.Duration,
			Distance: r.Distance,
		})
	}
	c.debugf("activities: %d match date %s", len(activities), dateStr)
	return activities, nil
}

var csrfRe = regexp.MustCompile(`name="_csrf"\s+value="([^"]+)"`)
var appCSRFRe = regexp.MustCompile(`<meta\s+name="csrf-token"\s+content="([^"]+)"`)
var ticketRe = regexp.MustCompile(`ticket=([A-Za-z0-9_\-]+)`)

func extractCSRF(body []byte) string {
	m := csrfRe.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}

func extractAppCSRF(body []byte) string {
	m := appCSRFRe.FindSubmatch(body)
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
