package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// User agent for HTTP requests
	UserAgent string

	// Monitoring settings
	Subreddit    string
	PollInterval time.Duration
	PostLimit    int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		UserAgent: os.Getenv("USER_AGENT"),
		Subreddit: os.Getenv("SUBREDDIT"),
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

	return cfg, nil
}
