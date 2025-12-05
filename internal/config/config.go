package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	// User agent for HTTP requests
	UserAgent string

	// Monitoring settings
	Subreddits   []string
	PollInterval time.Duration
	PostLimit    int

	// Discord integration
	DiscordWebhookURL string

	// Claude AI integration
	AnthropicAPIKey string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		UserAgent:         os.Getenv("USER_AGENT"),
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		AnthropicAPIKey:   os.Getenv("ANTHROPIC_API_KEY"),
	}

	// Parse comma-separated subreddits
	subredditStr := os.Getenv("SUBREDDIT")
	if subredditStr == "" {
		cfg.Subreddits = []string{"poker"}
	} else {
		// Split by comma and trim whitespace
		subreddits := strings.Split(subredditStr, ",")
		for i, s := range subreddits {
			subreddits[i] = strings.TrimSpace(s)
		}
		cfg.Subreddits = subreddits
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = "RedditPokerBot/1.0"
	}

	// Parse poll interval
	pollIntervalStr := os.Getenv("POLL_INTERVAL_SECONDS")
	if pollIntervalStr == "" {
		cfg.PollInterval = 30 * time.Second
	} else {
		seconds, err := strconv.Atoi(pollIntervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid POLL_INTERVAL_SECONDS: %w", err)
		}
		cfg.PollInterval = time.Duration(seconds) * time.Second
	}

	// Parse post limit
	postLimitStr := os.Getenv("POST_LIMIT")
	if postLimitStr == "" {
		cfg.PostLimit = 25
	} else {
		limit, err := strconv.Atoi(postLimitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid POST_LIMIT: %w", err)
		}
		cfg.PostLimit = limit
	}

	return cfg, nil
}
