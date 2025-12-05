package reddit

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// PostHandler is called when a new post is detected
type PostHandler func(post Post)

// Monitor watches a subreddit for new posts
type Monitor struct {
	client       *Client
	subreddit    string
	pollInterval time.Duration
	postLimit    int
	handler      PostHandler

	seenPosts map[string]bool
	mu        sync.RWMutex
	startTime time.Time
}

// NewMonitor creates a new subreddit monitor
func NewMonitor(client *Client, subreddit string, pollInterval time.Duration, postLimit int, handler PostHandler) *Monitor {
	return &Monitor{
		client:       client,
		subreddit:    subreddit,
		pollInterval: pollInterval,
		postLimit:    postLimit,
		handler:      handler,
		seenPosts:    make(map[string]bool),
		startTime:    time.Now(),
	}
}

// Start begins monitoring the subreddit
func (m *Monitor) Start(ctx context.Context) error {
	log.Printf("Starting monitor for r/%s (polling every %s)", m.subreddit, m.pollInterval)

	// Do initial fetch to populate seen posts
	if err := m.initialFetch(ctx); err != nil {
		return fmt.Errorf("initial fetch failed: %w", err)
	}

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Monitor stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := m.poll(ctx); err != nil {
				log.Printf("Poll error: %v", err)
				// Continue polling despite errors
			}
		}
	}
}

// initialFetch gets the current posts to establish a baseline
func (m *Monitor) initialFetch(ctx context.Context) error {
	posts, err := m.client.GetNewPosts(ctx, m.subreddit, m.postLimit)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, post := range posts {
		m.seenPosts[post.ID] = true
	}

	log.Printf("Initial fetch complete: %d posts tracked", len(posts))
	return nil
}

// poll checks for new posts
func (m *Monitor) poll(ctx context.Context) error {
	log.Printf("[r/%s] Polling for new posts...", m.subreddit)

	posts, err := m.client.GetNewPosts(ctx, m.subreddit, m.postLimit)
	if err != nil {
		return err
	}

	var newPosts []Post

	m.mu.Lock()
	for _, post := range posts {
		if !m.seenPosts[post.ID] {
			m.seenPosts[post.ID] = true
			// Only notify for posts created after monitor started
			if post.CreatedAt.After(m.startTime) {
				newPosts = append(newPosts, post)
			}
		}
	}
	m.mu.Unlock()

	// Call handler for each new post (in reverse order so oldest first)
	for i := len(newPosts) - 1; i >= 0; i-- {
		m.handler(newPosts[i])
	}

	if len(newPosts) > 0 {
		log.Printf("[r/%s] Found %d new post(s)", m.subreddit, len(newPosts))
	} else {
		log.Printf("[r/%s] No new posts", m.subreddit)
	}

	return nil
}

// GetSeenCount returns the number of posts that have been seen
func (m *Monitor) GetSeenCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.seenPosts)
}
