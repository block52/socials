package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Reddit OAuth2 credentials
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	UserAgent    string

	// Monitoring settings
	Subreddit     string
	PollInterval  time.Duration
	PostLimit     int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		ClientID:     os.Getenv("REDDIT_CLIENT_ID"),
		ClientSecret: os.Getenv("REDDIT_CLIENT_SECRET"),
		Username:     os.Getenv("REDDIT_USERNAME"),
		Password:     os.Getenv("REDDIT_PASSWORD"),
		UserAgent:    os.Getenv("REDDIT_USER_AGENT"),
		Subreddit:    os.Getenv("REDDIT_SUBREDDIT"),
	}

	// Set defaults
	if cfg.Subreddit == "" {
		cfg.Subreddit = "poker"
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

	// Validate required fields
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("REDDIT_CLIENT_ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("REDDIT_CLIENT_SECRET is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("REDDIT_USERNAME is required")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("REDDIT_PASSWORD is required")
	}

	return cfg, nil
}
