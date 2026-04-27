package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://slack.com/api"

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

func New(token string) *Client {
	return NewWithBaseURL(token, defaultBaseURL)
}

func NewWithBaseURL(token, baseURL string) *Client {
	return &Client{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{},
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
