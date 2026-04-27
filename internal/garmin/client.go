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

// errSessionExpired is returned by fetchActivities when the server responds
// with 401/403, indicating the cached session needs to be refreshed.
var errSessionExpired = errors.New("garmin session expired")

// Client authenticates with Garmin Connect and fetches activities.
type Client struct {
	email       string
	password    string
	ssoBase     string
	connectBase string
	http        *http.Client
	debug       io.Writer
	appCSRF     string // CSRF token from the Garmin Connect app page
	jwtFGP      string // JWT fingerprint cookie (captured but currently unused)
	sessionFile string // path to session cache file; empty disables caching
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

// FetchActivities returns activities for the given date.
// It tries a cached session first; if the session is expired it re-authenticates
// once and saves the new session before retrying.
func (c *Client) FetchActivities(date time.Time) ([]Activity, error) {
	if c.sessionFile != "" && c.loadSession() == nil {
		activities, err := c.fetchActivities(date)
		if err == nil {
			return activities, nil
		}
		if errors.Is(err, errSessionExpired) {
			c.debugf("cached session expired — clearing cache and re-authenticating")
			c.resetJar()
			c.appCSRF = ""
		} else {
			return nil, err
		}
	}

	if err := c.authenticate(); err != nil {
		return nil, err
	}
	c.saveSession()
	return c.fetchActivities(date)
}

func (c *Client) resetJar() {
	jar, _ := cookiejar.New(nil)
	c.http.Jar = jar
}

func (c *Client) authenticate() error {
	signinURL := c.ssoBase + "/sso/signin"
	// Use /app/ as the service URL — Garmin's new Connect app lives there.
	// Requesting /modern/ causes a re-authentication loop through the SSO portal.
	params := url.Values{
		"service":   {c.connectBase + "/app/"},
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

	// Omit embed=true so Garmin issues a redirect (not an embedded ticket in the body).
	// Go's http.Client follows the 302 → /app/?ticket=... automatically, setting session cookies.
	form := url.Values{
		"username":  {c.email},
		"password":  {c.password},
		"_csrf":     {csrf},
		"service":   {c.connectBase + "/app/"},
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

	// Garmin SSO redirects POST → 302 → /app/?ticket=ST-XXX.
	// Go follows the redirect automatically; we capture the ticket URL to detect success.
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
		c.ensureAppCSRF()
		return nil
	}

	// Fallback: some SSO variants still embed the ticket in the body.
	ticket := extractTicket(body2)
	if ticket == "" {
		c.debugf("ticket: not found (neither in redirect nor body) — status %d", resp2.StatusCode)
		return errors.New("garmin login failed — check credentials")
	}
	c.debugf("ticket: found in response body (fallback flow)")

	resp3, err := c.http.Get(c.connectBase + "/app/?ticket=" + ticket)
	if err != nil {
		return fmt.Errorf("garmin login failed: %w", err)
	}
	appBody, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	c.appCSRF = extractAppCSRF(appBody)
	c.captureJWTFGP()
	c.debugf("auth complete via fallback/body flow (app CSRF: %v, JWT_FGP: %v)", c.appCSRF != "", c.jwtFGP != "")
	c.ensureAppCSRF()
	return nil
}

// ensureAppCSRF fetches /app/ explicitly if the CSRF token wasn't found in the
// auth redirect body. Garmin may land on /modern/ first and redirect to /app/
// without embedding the token in the intermediate page.
func (c *Client) ensureAppCSRF() {
	if c.appCSRF != "" {
		return
	}
	c.debugf("CSRF token not found in auth response — fetching /app/ explicitly")
	resp, err := c.http.Get(c.connectBase + "/app/")
	if err != nil {
		c.debugf("ensureAppCSRF: GET /app/ error: %v", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	c.appCSRF = extractAppCSRF(body)
	c.captureJWTFGP()
	c.debugf("ensureAppCSRF: CSRF found=%v", c.appCSRF != "")
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
	req.Header.Set("Origin", c.connectBase)
	req.Header.Set("Referer", c.connectBase+"/app/")
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

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		c.debugf("activities: HTTP %d — session expired", resp.StatusCode)
		return nil, errSessionExpired
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
