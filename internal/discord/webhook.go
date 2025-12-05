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
	Content    string  `json:"content,omitempty"`
	Embeds     []Embed `json:"embeds,omitempty"`
	ThreadName string  `json:"thread_name,omitempty"`
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

// SendPostWithResponses sends a Reddit post to Discord with suggested responses in a thread
func (c *WebhookClient) SendPostWithResponses(ctx context.Context, post reddit.Post, witty, formal, block52 string) error {
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

	// Create thread name from post title (max 100 chars)
	threadName := post.Title
	if len(threadName) > 97 {
		threadName = threadName[:97] + "..."
	}

	msg := WebhookMessage{
		Embeds:     []Embed{embed},
		ThreadName: "Response Suggestions: " + threadName,
	}

	// Send initial message with thread
	if err := c.send(ctx, msg); err != nil {
		return err
	}

	// Note: To post the responses in the thread, we'd need to use the ?thread_id parameter
	// For now, we'll send them as a follow-up message in the thread
	// This requires getting the thread ID from the first message response

	// Create a follow-up message with the responses
	responseContent := fmt.Sprintf("**Suggested Responses:**\n\n**1️⃣ Witty:**\n%s\n\n**2️⃣ Formal:**\n%s\n\n**3️⃣ Block52 Angle:**\n%s",
		witty, formal, block52)

	followUpMsg := WebhookMessage{
		Content: responseContent,
	}

	// Wait a moment for the thread to be created
	time.Sleep(500 * time.Millisecond)

	// Send follow-up (this will go in the main channel for now)
	// To properly thread, we'd need Discord API, not just webhooks
	return c.send(ctx, followUpMsg)
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
