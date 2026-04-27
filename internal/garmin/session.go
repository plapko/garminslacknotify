package garmin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type sessionCache struct {
	SavedAt string          `json:"saved_at"`
	AppCSRF string          `json:"app_csrf"`
	Cookies []sessionCookie `json:"cookies"`
}

type sessionCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SetSessionFile enables persistent session caching. The client will load
// cookies from path before authenticating and save them after a successful
// login, so most cron runs skip the SSO flow entirely.
func (c *Client) SetSessionFile(path string) *Client {
	c.sessionFile = path
	return c
}

func (c *Client) saveSession() {
	if c.sessionFile == "" {
		return
	}

	// Collect cookies visible at /app/ (covers Path=/ and Path=/app/).
	appURL, _ := url.Parse(c.connectBase + "/app/")
	seen := map[string]bool{}
	var cookies []sessionCookie
	for _, ck := range c.http.Jar.Cookies(appURL) {
		if !seen[ck.Name] {
			seen[ck.Name] = true
			cookies = append(cookies, sessionCookie{Name: ck.Name, Value: ck.Value})
		}
	}

	cache := sessionCache{
		SavedAt: time.Now().UTC().Format(time.RFC3339),
		AppCSRF: c.appCSRF,
		Cookies: cookies,
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		c.debugf("session: marshal failed: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(c.sessionFile), 0o700); err != nil {
		c.debugf("session: mkdir failed: %v", err)
		return
	}
	if err := os.WriteFile(c.sessionFile, data, 0o600); err != nil {
		c.debugf("session: write failed: %v", err)
		return
	}
	c.debugf("session: saved %d cookies to %s", len(cookies), c.sessionFile)
}

func (c *Client) loadSession() error {
	if c.sessionFile == "" {
		return os.ErrNotExist
	}
	data, err := os.ReadFile(c.sessionFile)
	if err != nil {
		return err
	}
	var cache sessionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return err
	}
	if cache.AppCSRF == "" || len(cache.Cookies) == 0 {
		return errors.New("incomplete session cache")
	}

	// Inject cookies at root so they apply to /gc-api/, /app/, etc.
	rootURL, _ := url.Parse(c.connectBase + "/")
	httpCookies := make([]*http.Cookie, len(cache.Cookies))
	for i, ck := range cache.Cookies {
		httpCookies[i] = &http.Cookie{Name: ck.Name, Value: ck.Value}
	}
	c.http.Jar.SetCookies(rootURL, httpCookies)
	c.appCSRF = cache.AppCSRF
	c.debugf("session: loaded from %s (saved %s, %d cookies)", c.sessionFile, cache.SavedAt, len(cache.Cookies))
	return nil
}
