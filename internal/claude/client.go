package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/block52/reddit-poker-bot/internal/reddit"
)

const (
	apiURL     = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
	model      = "claude-3-5-sonnet-20241022"
)

// Client interacts with Claude API
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// Response represents suggested responses to a post
type Response struct {
	Witty      string
	Formal     string
	Block52    string
}

// NewClient creates a new Claude API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateResponses generates three response suggestions for a Reddit post
func (c *Client) GenerateResponses(ctx context.Context, post reddit.Post) (*Response, error) {
	if c.apiKey == "" {
		return nil, nil // Claude integration disabled
	}

	prompt := fmt.Sprintf(`You are helping generate response suggestions for a Reddit post in r/%s.

Post Title: %s
Post Link: %s

Generate exactly 3 different response comments for this post:

1. WITTY: A clever, witty response with personality and humor (but not offensive)
2. FORMAL: A more professional, helpful, and informative response
3. BLOCK52: A response that naturally mentions how Block52 (a blockchain poker platform) could be relevant or helpful here. Only include this if it's genuinely applicable - if Block52 isn't relevant to this post, say "N/A"

Format your response as JSON with these exact keys: "witty", "formal", "block52"
Keep each response under 280 characters.`, post.Subreddit, post.Title, post.GetFullPermalink())

	resp, err := c.callAPI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// callAPI makes a request to Claude API
func (c *Client) callAPI(ctx context.Context, prompt string) (*Response, error) {
	reqBody := map[string]interface{}{
		"model": model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("API request failed with status %d: %v", resp.StatusCode, errResp)
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Parse the JSON response from Claude
	var responses Response
	text := apiResp.Content[0].Text

	// Try to extract JSON from the response
	if err := json.Unmarshal([]byte(text), &responses); err != nil {
		// If parsing fails, return a structured error
		return nil, fmt.Errorf("parsing response JSON: %w", err)
	}

	return &responses, nil
}
