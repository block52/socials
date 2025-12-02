package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	authURL    = "https://www.reddit.com/api/v1/access_token"
	apiBaseURL = "https://oauth.reddit.com"
)

// Client is a Reddit API client
type Client struct {
	clientID     string
	clientSecret string
	username     string
	password     string
	userAgent    string

	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
	mu          sync.RWMutex
}

// Post represents a Reddit post
type Post struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Subreddit   string    `json:"subreddit"`
	URL         string    `json:"url"`
	Permalink   string    `json:"permalink"`
	SelfText    string    `json:"selftext"`
	Score       int       `json:"score"`
	NumComments int       `json:"num_comments"`
	CreatedUTC  float64   `json:"created_utc"`
	CreatedAt   time.Time `json:"-"`
	Flair       string    `json:"link_flair_text"`
	IsNSFW      bool      `json:"over_18"`
	IsSelf      bool      `json:"is_self"`
}

// ListingResponse represents a Reddit API listing response
type ListingResponse struct {
	Kind string `json:"kind"`
	Data struct {
		After    string `json:"after"`
		Before   string `json:"before"`
		Children []struct {
			Kind string `json:"kind"`
			Data Post   `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

// TokenResponse represents the OAuth2 token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// NewClient creates a new Reddit API client
func NewClient(clientID, clientSecret, username, password, userAgent string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		userAgent:    userAgent,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// authenticate obtains an access token using password grant
func (c *Client) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if token is still valid (with 60 second buffer)
	if c.accessToken != "" && time.Now().Add(60*time.Second).Before(c.tokenExpiry) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}

	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// getAccessToken returns a valid access token, refreshing if necessary
func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	if err := c.authenticate(ctx); err != nil {
		return "", err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken, nil
}

// GetNewPosts fetches the newest posts from a subreddit
func (c *Client) GetNewPosts(ctx context.Context, subreddit string, limit int) ([]Post, error) {
	token, err := c.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting access token: %w", err)
	}

	endpoint := fmt.Sprintf("%s/r/%s/new?limit=%d", apiBaseURL, subreddit, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var listing ListingResponse
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	posts := make([]Post, 0, len(listing.Data.Children))
	for _, child := range listing.Data.Children {
		post := child.Data
		post.CreatedAt = time.Unix(int64(post.CreatedUTC), 0)
		posts = append(posts, post)
	}

	return posts, nil
}

// GetFullPermalink returns the full Reddit URL for a post
func (p *Post) GetFullPermalink() string {
	return "https://reddit.com" + p.Permalink
}
