package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/block52/reddit-poker-bot/internal/reddit"
)

// WebhookClient sends messages to Discord via webhook
type WebhookClient struct {
	webhookURL string
	httpClient *http.Client
}

// WebhookMessage represents a Discord webhook message
type WebhookMessage struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

// Embed represents a Discord embed
type Embed struct {
	Title       string    `json:"title,omitempty"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Color       int       `json:"color,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Footer      *Footer   `json:"footer,omitempty"`
	Author      *Author   `json:"author,omitempty"`
	Fields      []Field   `json:"fields,omitempty"`
}

// Footer represents an embed footer
type Footer struct {
	Text string `json:"text,omitempty"`
}

// Author represents an embed author
type Author struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Field represents an embed field
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// NewWebhookClient creates a new Discord webhook client
func NewWebhookClient(webhookURL string) *WebhookClient {
	return &WebhookClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendPost sends a Reddit post to Discord as an embed
func (c *WebhookClient) SendPost(ctx context.Context, post reddit.Post) error {
	if c.webhookURL == "" {
		return nil // Discord integration disabled
	}

	age := time.Since(post.CreatedAt)
	ageStr := formatDuration(age)

	embed := Embed{
		Title:     post.Title,
		URL:       post.GetFullPermalink(),
		Color:     0xFF4500, // Reddit orange
		Timestamp: post.CreatedAt,
		Author: &Author{
			Name: fmt.Sprintf("u/%s", post.Author),
			URL:  fmt.Sprintf("https://reddit.com/u/%s", post.Author),
		},
		Footer: &Footer{
			Text: fmt.Sprintf("r/%s • Posted %s ago", post.Subreddit, ageStr),
		},
	}

	msg := WebhookMessage{
		Embeds: []Embed{embed},
	}

	return c.send(ctx, msg)
}

// send sends a webhook message to Discord
func (c *WebhookClient) send(ctx context.Context, msg WebhookMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling webhook message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request failed with status %d", resp.StatusCode)
	}

	return nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
