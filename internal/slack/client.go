package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/plapko/garminslacknotify/internal/httpdebug"
)

const defaultBaseURL = "https://slack.com/api"

type Client struct {
	token   string
	baseURL string
	http    *http.Client
	debug   io.Writer
}

func New(token string) *Client {
	return build(token, defaultBaseURL, nil)
}

func NewWithBaseURL(token, baseURL string) *Client {
	return build(token, baseURL, nil)
}

// NewWithDebug creates a Client that writes HTTP debug info to debug.
func NewWithDebug(token string, debug io.Writer) *Client {
	return build(token, defaultBaseURL, debug)
}

func build(token, baseURL string, debug io.Writer) *Client {
	var transport http.RoundTripper = http.DefaultTransport
	if debug != nil {
		transport = &httpdebug.Transport{Base: transport, Out: debug, Label: "slack"}
	}
	return &Client{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{Transport: transport},
		debug:   debug,
	}
}

func (c *Client) SetStatus(text, emoji string) error {
	payload := map[string]interface{}{
		"profile": map[string]interface{}{
			"status_text":       text,
			"status_emoji":      ":" + emoji + ":",
			"status_expiration": 0,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if c.debug != nil {
		fmt.Fprintf(c.debug, "[debug] slack   setting status %q  emoji :%s:\n", text, emoji)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/users.profile.set", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if !result.OK {
		return fmt.Errorf("slack API error: %s", result.Error)
	}
	return nil
}
