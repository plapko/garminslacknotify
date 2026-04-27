package httpdebug

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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
	fmt.Fprintf(t.Out, "[debug] %s → %s %s\n", t.Label, req.Method, req.URL)

	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(t.Out, "[debug] %s ← error: %v\n\n", t.Label, err)
		return nil, err
	}

	body, readErr := io.ReadAll(resp.Body)
	resp.Body.Close()

	snippet := string(body)
	if len(snippet) > snippetLen {
		snippet = snippet[:snippetLen] + "…"
	}
	fmt.Fprintf(t.Out, "[debug] %s ← %s  (%d bytes)\n%s\n\n",
		t.Label, resp.Status, len(body), indent(snippet))

	if readErr != nil {
		return nil, readErr
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

func indent(s string) string {
	if s == "" {
		return "        (empty body)"
	}
	return "        " + s
}
