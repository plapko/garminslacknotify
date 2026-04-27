package httpdebug

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const snippetLen = 500

// Transport wraps an http.RoundTripper and writes each request/response to Out.
// Request bodies are never logged (may contain credentials).
// Response bodies are captured, logged (truncated to snippetLen bytes), and
// replaced with a fresh reader so callers can still read them normally.
type Transport struct {
	Base  http.RoundTripper
	Out   io.Writer
	Label string
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log outgoing cookies so we can verify the session is being sent.
	var cookieNames []string
	for _, c := range req.Cookies() {
		cookieNames = append(cookieNames, c.Name)
	}
	cookieSummary := "(none)"
	if len(cookieNames) > 0 {
		cookieSummary = strings.Join(cookieNames, ", ")
	}

	// Log key request headers so we can verify auth/CSRF headers are present.
	var hdrs []string
	for _, name := range []string{
		"connect-csrf-token", "Origin", "Referer",
		"X-Requested-With", "Authorization",
	} {
		if v := req.Header.Get(name); v != "" {
			if len(v) > 24 {
				v = v[:24] + "…"
			}
			hdrs = append(hdrs, name+": "+v)
		}
	}
	hdrLine := ""
	if len(hdrs) > 0 {
		hdrLine = "\n        req-hdrs: " + strings.Join(hdrs, " | ")
	}

	fmt.Fprintf(t.Out, "[debug] %s → %s %s\n        cookies: %s%s\n",
		t.Label, req.Method, req.URL, cookieSummary, hdrLine)

	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(t.Out, "[debug] %s ← error: %v\n\n", t.Label, err)
		return nil, err
	}

	// Log Set-Cookie headers so we can track session establishment.
	var setCookies []string
	for _, sc := range resp.Cookies() {
		setCookies = append(setCookies, sc.Name+"="+sc.Value[:min(len(sc.Value), 12)]+"…")
	}
	setCookieLine := ""
	if len(setCookies) > 0 {
		setCookieLine = "\n        set-cookie: " + strings.Join(setCookies, ", ")
	}

	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()

	snippet := string(body)
	if len(snippet) > snippetLen {
		snippet = snippet[:snippetLen] + "…"
	}
	fmt.Fprintf(t.Out, "[debug] %s ← %s  (%d bytes)%s\n%s\n\n",
		t.Label, resp.Status, len(body), setCookieLine, indent(snippet))

	if readErr != nil {
		return nil, readErr
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func indent(s string) string {
	if s == "" {
		return "        (empty body)"
	}
	return "        " + s
}
